package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xiaozhi-go-core/internal/config"
)

func TestVersionGreater(t *testing.T) {
	if !versionGreater("1.2.10", "1.2.9") {
		t.Fatal("expected 1.2.10 > 1.2.9")
	}
	if versionGreater("1.2.0", "1.2.1") {
		t.Fatal("expected 1.2.0 <= 1.2.1")
	}
}

func TestOTAWebsocketResponse(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gateway.WebSocketURL = "ws://example.com/xiaozhi/v1/"
	cfg.Gateway.Auth.Enabled = true
	cfg.Gateway.AuthKey = "secret"
	cfg.OTA.BinDir = t.TempDir()
	handler := NewOTAHandler(config.NewStore(cfg), slog.Default())

	req := httptest.NewRequest(http.MethodPost, "http://ota.example.com/xiaozhi/ota/", strings.NewReader(`{"application":{"version":"1.0.0"}}`))
	req.Header.Set("device-id", "device-1")
	req.Header.Set("client-id", "client-1")
	rec := httptest.NewRecorder()

	handler.handleOTA(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	ws, ok := payload["websocket"].(map[string]any)
	if !ok {
		t.Fatalf("missing websocket payload: %#v", payload)
	}
	if ws["url"] != "ws://example.com/xiaozhi/v1/" {
		t.Fatalf("unexpected websocket url: %v", ws["url"])
	}
	if ws["token"] == "" {
		t.Fatal("expected auth token")
	}
}
