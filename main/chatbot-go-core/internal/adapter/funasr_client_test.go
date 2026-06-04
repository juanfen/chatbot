package adapter

import "testing"

func TestFunASRPrefixCleanup(t *testing.T) {
	input := "<|zh|><|NEUTRAL|><|Speech|>你好，小智"
	got := funASRPrefixPattern.ReplaceAllString(input, "")
	if got != "你好，小智" {
		t.Fatalf("unexpected cleanup result: %q", got)
	}
}

func TestPCM16ToWAVHeader(t *testing.T) {
	wav := PCM16ToWAV([]byte{1, 0, 2, 0}, 16000, 1)
	if len(wav) != 48 {
		t.Fatalf("unexpected wav length: %d", len(wav))
	}
	if string(wav[:4]) != "RIFF" {
		t.Fatalf("missing RIFF header")
	}
	if string(wav[8:12]) != "WAVE" {
		t.Fatalf("missing WAVE header")
	}
}
