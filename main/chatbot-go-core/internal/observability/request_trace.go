package observability

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type RequestTrace struct {
	mu   sync.Mutex
	file *os.File
}

func NewRequestTrace(requestID string) (*RequestTrace, error) {
	safeID := sanitizeFileName(strings.TrimSpace(requestID))
	if safeID == "" {
		safeID = "unknown-request"
	}
	dir := filepath.Join("logs", "requests")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, safeID+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &RequestTrace{file: f}, nil
}

func (t *RequestTrace) Log(stage string, kv ...any) {
	if t == nil || t.file == nil {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	var b strings.Builder
	b.WriteString(ts)
	b.WriteString(" stage=")
	b.WriteString(stage)
	for i := 0; i+1 < len(kv); i += 2 {
		k := fmt.Sprint(kv[i])
		v := fmt.Sprint(kv[i+1])
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	b.WriteByte('\n')

	t.mu.Lock()
	defer t.mu.Unlock()
	_, _ = t.file.WriteString(b.String())
}

func (t *RequestTrace) Close() {
	if t == nil || t.file == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	_ = t.file.Close()
	t.file = nil
}

func sanitizeFileName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
