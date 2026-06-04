package service

import (
	"context"
	"testing"
	"time"
)

func TestSlidingCtxTouchExtendsIdle(t *testing.T) {
	sliding, ctx := newSlidingCtx(context.Background(), 200*time.Millisecond)
	defer sliding.stop()

	d1, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}

	time.Sleep(120 * time.Millisecond)
	sliding.touch()

	d2, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline after touch")
	}
	if !d2.After(d1) {
		t.Fatalf("deadline should extend after touch: d1=%v d2=%v", d1, d2)
	}

	time.Sleep(250 * time.Millisecond)
	if ctx.Err() == nil {
		t.Fatal("expected idle timeout after touch window")
	}
}

func TestStreamContextHardCap(t *testing.T) {
	ctx, _, stop := streamContext(80, 500)
	defer stop()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	if time.Until(deadline) > 100*time.Millisecond {
		t.Fatalf("hard cap should bind before idle: %v", time.Until(deadline))
	}
	time.Sleep(90 * time.Millisecond)
	if ctx.Err() == nil {
		t.Fatal("expected hard timeout")
	}
}

func TestSlidingCtxTouchDoesNotPanicUnderConcurrency(t *testing.T) {
	sliding, ctx := newSlidingCtx(context.Background(), 2*time.Second)
	defer sliding.stop()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 200; i++ {
			sliding.touch()
			_ = ctx.Err()
			time.Sleep(time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("touch loop hung")
	}
}
