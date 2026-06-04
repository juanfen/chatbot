package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

func NewLogger() *slog.Logger {
	level := parseLogLevel(os.Getenv("XIAOZHI_GO_LOG_LEVEL"))
	if strings.EqualFold(os.Getenv("XIAOZHI_GO_LOG_FORMAT"), "json") {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String(slog.TimeKey, time.Now().Format("2006-01-02 15:04:05"))
				}
				return a
			},
		}))
	}
	return slog.New(newPrettyHandler(os.Stdout, level))
}

type prettyHandler struct {
	out    io.Writer
	level  slog.Leveler
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func newPrettyHandler(out io.Writer, level slog.Leveler) *prettyHandler {
	return &prettyHandler{
		out:   out,
		level: level,
		mu:    &sync.Mutex{},
	}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	t := r.Time
	if t.IsZero() {
		t = time.Now()
	}
	b.WriteString(t.Format("01-02 15:04:05.000"))
	b.WriteByte(' ')
	b.WriteString(formatLevel(r.Level))
	b.WriteByte(' ')
	b.WriteString(r.Message)

	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	for _, a := range attrs {
		h.appendAttr(&b, a, h.groups)
	}
	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := h.clone()
	cp.attrs = append(cp.attrs, attrs...)
	return cp
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	cp := h.clone()
	if name != "" {
		cp.groups = append(cp.groups, name)
	}
	return cp
}

func (h *prettyHandler) clone() *prettyHandler {
	cp := *h
	cp.attrs = append([]slog.Attr(nil), h.attrs...)
	cp.groups = append([]string(nil), h.groups...)
	return &cp
}

func (h *prettyHandler) appendAttr(b *strings.Builder, a slog.Attr, groups []string) {
	if a.Key == "" {
		return
	}
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		nextGroups := append(append([]string(nil), groups...), a.Key)
		for _, child := range a.Value.Group() {
			h.appendAttr(b, child, nextGroups)
		}
		return
	}
	value := attrValueString(a.Value)
	if value == "" {
		return
	}
	key := friendlyKey(a.Key)
	if len(groups) > 0 {
		key = strings.Join(append(groups, key), ".")
	}
	b.WriteByte(' ')
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(quoteIfNeeded(value))
}

func attrValueString(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format("15:04:05.000")
	default:
		return fmt.Sprint(v.Any())
	}
}

func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t\n\r=:") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

func friendlyKey(key string) string {
	switch key {
	case "session_id":
		return "session"
	case "request_id":
		return "request"
	case "device_id":
		return "device"
	case "client_id":
		return "client"
	case "elapsed_ms":
		return "elapsed"
	case "error_message":
		return "error"
	default:
		return key
	}
}

func formatLevel(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN "
	case level >= slog.LevelInfo:
		return "INFO "
	default:
		return "DEBUG"
	}
}

func parseLogLevel(raw string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
