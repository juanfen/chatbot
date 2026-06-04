package adapter

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"xiaozhi-go-core/internal/config"
)

func TestDeviceAudioClientUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/device/audio/upload" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("mac_id") != "AA:BB:CC:DD:EE:FF" {
			t.Fatalf("mac_id: %q", r.FormValue("mac_id"))
		}
		if r.FormValue("audio_len") != "5.0s" {
			t.Fatalf("audio_len: %q", r.FormValue("audio_len"))
		}
		if r.FormValue("context") != "你好世界" {
			t.Fatalf("context: %q", r.FormValue("context"))
		}
		file, _, err := r.FormFile("audio_blob")
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "RIFF") {
			t.Fatalf("expected wav header, got %q", data[:4])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "ok",
			"data": map[string]any{
				"id":         "rec-1",
				"macId":      "AA:BB:CC:DD:EE:FF",
				"audioLen":   "5.0s",
				"context":    "你好世界",
				"createTime": "2026-06-03T10:00:00+08:00",
			},
		})
	}))
	defer srv.Close()

	host, portStr, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.ChatbotAdmin.Enabled = true
	cfg.ChatbotAdmin.IP = host
	if host == "::1" {
		cfg.ChatbotAdmin.IP = "127.0.0.1"
	}
	cfg.ChatbotAdmin.Port = port

	store := config.NewStore(cfg)
	client := NewDeviceAudioClient(store, nil)
	wav := PCM16ToWAV([]byte{0, 0, 1, 0, 2, 0}, 16000, 1)
	result, err := client.Upload(context.Background(), "AA:BB:CC:DD:EE:FF", "5.0s", "你好世界", wav)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "rec-1" {
		t.Fatalf("id: %q", result.ID)
	}
}
