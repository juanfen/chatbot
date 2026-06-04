package gateway

import (
	"context"
	"sync"
	"time"
)

type audioQueueItem struct {
	kind       string
	data       []byte
	durationMs int
	fn         func() error
}

// AudioRateController 按每包真实音频时长（durationMs）相对墙钟节流下发，避免固定 sendDelay 导致发送过快。
type AudioRateController struct {
	frameDuration time.Duration
	minSendGap    time.Duration
	sendAudio     func([]byte) error
	mu            sync.Mutex
	cond          *sync.Cond
	queue         []audioQueueItem
	active        bool
	playPosition  time.Duration
	startTime     time.Time
	done          chan struct{}
	cancel        context.CancelFunc
}

func NewAudioRateController(frameDurationMs int, sendDelayMs int, sendAudio func([]byte) error) *AudioRateController {
	if frameDurationMs <= 0 {
		frameDurationMs = 60
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &AudioRateController{
		frameDuration: time.Duration(frameDurationMs) * time.Millisecond,
		sendAudio:     sendAudio,
		done:          make(chan struct{}),
		cancel:        cancel,
	}
	if sendDelayMs > 0 {
		c.minSendGap = time.Duration(sendDelayMs) * time.Millisecond
	}
	c.cond = sync.NewCond(&c.mu)
	go c.loop(ctx)
	return c
}

// Reset 在新的 TTS 播放开始前重置时间轴与队列。
func (c *AudioRateController) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queue = nil
	c.playPosition = 0
	c.startTime = time.Time{}
	c.active = false
}

func (c *AudioRateController) AddAudio(packet []byte, durationMs int) error {
	if len(packet) == 0 {
		return nil
	}
	if durationMs <= 0 {
		durationMs = int(c.frameDuration / time.Millisecond)
	}
	c.mu.Lock()
	c.queue = append(c.queue, audioQueueItem{
		kind:       "audio",
		data:       append([]byte(nil), packet...),
		durationMs: durationMs,
	})
	c.cond.Signal()
	c.mu.Unlock()
	return nil
}

func (c *AudioRateController) AddMessage(fn func() error) {
	c.mu.Lock()
	c.queue = append(c.queue, audioQueueItem{kind: "message", fn: fn})
	c.cond.Signal()
	c.mu.Unlock()
}

func (c *AudioRateController) WaitEmpty() {
	for {
		c.mu.Lock()
		empty := len(c.queue) == 0 && !c.active
		var tailWait time.Duration
		if empty && !c.startTime.IsZero() {
			elapsed := time.Since(c.startTime)
			if elapsed < c.playPosition {
				tailWait = c.playPosition - elapsed
			}
		}
		c.mu.Unlock()
		if empty {
			if tailWait > 0 {
				time.Sleep(tailWait)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *AudioRateController) Stop() {
	c.cancel()
	c.cond.Broadcast()
	<-c.done
}

func (c *AudioRateController) loop(ctx context.Context) {
	defer close(c.done)
	for {
		c.mu.Lock()
		for len(c.queue) == 0 {
			c.cond.Wait()
			select {
			case <-ctx.Done():
				c.mu.Unlock()
				return
			default:
			}
		}
		item := c.queue[0]
		c.queue = c.queue[1:]
		c.active = true
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		default:
		}

		if item.kind == "message" {
			if item.fn != nil {
				_ = item.fn()
			}
			c.mu.Lock()
			c.active = false
			c.mu.Unlock()
			continue
		}

		frameDur := time.Duration(item.durationMs) * time.Millisecond
		if c.startTime.IsZero() {
			c.startTime = time.Now()
		}
		elapsed := time.Since(c.startTime)
		wait := time.Duration(0)
		if elapsed < c.playPosition {
			wait = c.playPosition - elapsed
		}
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		advance := frameDur
		if c.minSendGap > 0 && c.minSendGap > advance {
			advance = c.minSendGap
		}
		c.playPosition += advance
		_ = c.sendAudio(item.data)
		c.mu.Lock()
		c.active = false
		c.mu.Unlock()
	}
}
