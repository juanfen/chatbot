package adapter

import "testing"

func TestPCMS16FrameBytes24k60ms(t *testing.T) {
	got := PCMS16FrameBytes(24000, 1, 60)
	if got != 2880 {
		t.Fatalf("want 2880 bytes per frame, got %d", got)
	}
}

func TestEncodeS16LEFramesSplitsAndPads(t *testing.T) {
	enc, err := NewOpusEncoder(24000, 1)
	if err != nil {
		t.Fatal(err)
	}
	// 1.5 frames of PCM
	pcm := make([]byte, 2880+1440)
	for i := range pcm {
		pcm[i] = byte(i % 256)
	}
	frames, err := enc.EncodeS16LEFrames(pcm, 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 {
		t.Fatalf("want 2 opus frames (second padded), got %d", len(frames))
	}
	for i, f := range frames {
		if len(f) == 0 {
			t.Fatalf("frame %d empty", i)
		}
	}
}

func TestPCMS16DurationMs(t *testing.T) {
	if got := PCMS16DurationMs(2880, 24000, 1); got != 60 {
		t.Fatalf("want 60ms, got %d", got)
	}
}
