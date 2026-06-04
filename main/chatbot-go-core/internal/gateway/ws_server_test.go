package gateway

import (
	"os"
	"path/filepath"
	"testing"

	"xiaozhi-go-core/internal/config"
)

func TestNormalizeSpeechText(t *testing.T) {
	got := normalizeSpeechText("你，好，小智！")
	if got != "你好小智" {
		t.Fatalf("unexpected normalized text: %q", got)
	}
}

func TestTrimPunctuationOrEmoji(t *testing.T) {
	got := trimPunctuationOrEmoji("！你好🙂")
	if got != "你好" {
		t.Fatalf("unexpected trimmed text: %q", got)
	}
}

func TestAudioRateController(t *testing.T) {
	sent := make(chan []byte, 3)
	c := NewAudioRateController(1, -1, func(packet []byte) error {
		sent <- packet
		return nil
	})
	defer c.Stop()
	if err := c.AddAudio([]byte{1}, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.AddAudio([]byte{2}, 1); err != nil {
		t.Fatal(err)
	}
	c.WaitEmpty()
	if len(sent) != 2 {
		t.Fatalf("expected 2 packets, got %d", len(sent))
	}
}

func TestReadWakeupResponseYAML(t *testing.T) {
	dir := t.TempDir()
	audioPath := filepath.Join(dir, "wake.wav")
	if err := os.WriteFile(audioPath, make([]byte, 16*1024), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, ".wakeup_words.yaml")
	content := "voice_hash:\n  file_path: " + audioPath + "\n  text: 在的呢\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	path, text := readWakeupResponse(cfgPath, "fallback.wav")
	if path != audioPath || text != "在的呢" {
		t.Fatalf("unexpected wakeup response: %q %q", path, text)
	}
}

func TestDetectChatTextWakeup(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gateway.WakeupWords = []string{"你好小智"}
	cfg.Gateway.EnableGreeting = true
	session := &gatewaySession{store: config.NewStore(cfg)}
	if got := session.detectChatText("你好，小智！"); got != "嘿，你好呀" {
		t.Fatalf("unexpected wakeup response: %q", got)
	}
}

func TestMQTTHeader(t *testing.T) {
	session := &gatewaySession{}
	packet := session.withMQTTHeader([]byte{1, 2, 3})
	if len(packet) != 19 {
		t.Fatalf("unexpected packet length: %d", len(packet))
	}
	if packet[0] != 1 {
		t.Fatalf("unexpected packet type: %d", packet[0])
	}
	if packet[12] != 0 || packet[13] != 0 || packet[14] != 0 || packet[15] != 3 {
		t.Fatalf("unexpected opus length header: %#v", packet[12:16])
	}
}
