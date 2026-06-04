package service

import (
	"context"
	"sync"
	"time"
)

// slidingCtx 在 touch 时仅延长 idle 截止时间，不 cancel 内层 context，避免与 TTS/LLM 共用 ctx 冲突引发 panic。
type slidingCtx struct {
	parent  context.Context
	idleDur time.Duration

	mu      sync.Mutex
	idleEnd time.Time

	done       chan struct{}
	closeOnce  sync.Once
	watchStop  context.CancelFunc
}

func newSlidingCtx(parent context.Context, idle time.Duration) (*slidingCtx, context.Context) {
	if parent == nil {
		parent = context.Background()
	}
	if idle <= 0 {
		idle = 60 * time.Second
	}
	now := time.Now()
	s := &slidingCtx{
		parent:  parent,
		idleDur: idle,
		idleEnd: now.Add(idle),
		done:    make(chan struct{}),
	}
	watchCtx, watchStop := context.WithCancel(context.Background())
	s.watchStop = watchStop
	go s.watch(watchCtx)
	return s, s
}

func (s *slidingCtx) touch() {
	s.mu.Lock()
	s.idleEnd = time.Now().Add(s.idleDur)
	s.mu.Unlock()
}

func (s *slidingCtx) stop() {
	if s.watchStop != nil {
		s.watchStop()
	}
	s.fireDone()
}

func (s *slidingCtx) watch(watchCtx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-watchCtx.Done():
			return
		case <-s.parent.Done():
			s.fireDone()
			return
		case <-ticker.C:
			s.mu.Lock()
			expired := time.Now().After(s.idleEnd)
			s.mu.Unlock()
			if expired {
				s.fireDone()
				return
			}
		}
	}
}

func (s *slidingCtx) fireDone() {
	s.closeOnce.Do(func() { close(s.done) })
}

func (s *slidingCtx) Deadline() (time.Time, bool) {
	s.mu.Lock()
	idleEnd := s.idleEnd
	s.mu.Unlock()
	if pd, ok := s.parent.Deadline(); ok {
		if pd.Before(idleEnd) {
			return pd, true
		}
	}
	return idleEnd, true
}

func (s *slidingCtx) Done() <-chan struct{} {
	return s.done
}

func (s *slidingCtx) Err() error {
	select {
	case <-s.done:
		if err := s.parent.Err(); err != nil {
			return err
		}
		return context.DeadlineExceeded
	default:
		return nil
	}
}

func (s *slidingCtx) Value(key any) any {
	return s.parent.Value(key)
}

// streamContext 为单次 StreamChat 构造上下文：
// - hard：request_timeout_ms 绝对上限（<=0 表示不限制）
// - idle：stream_idle_timeout_ms，有活动时 touch 延长 idle 截止，不反复 cancel 子 context
func streamContext(requestTimeoutMs, streamIdleTimeoutMs int) (context.Context, func(), func()) {
	parent := context.Background()
	var hardCancel context.CancelFunc
	if requestTimeoutMs > 0 {
		parent, hardCancel = context.WithTimeout(parent, time.Duration(requestTimeoutMs)*time.Millisecond)
	}

	idleMs := streamIdleTimeoutMs
	if idleMs <= 0 {
		idleMs = requestTimeoutMs
	}
	if idleMs <= 0 {
		idleMs = 60000
	}

	sliding, ctx := newSlidingCtx(parent, time.Duration(idleMs)*time.Millisecond)
	stop := func() {
		sliding.stop()
		if hardCancel != nil {
			hardCancel()
		}
	}
	return ctx, sliding.touch, stop
}
