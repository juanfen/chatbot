package observability

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects in-process counters and histograms.
// A production system would export to Prometheus; this provides structured log output.
type Metrics struct {
	logger          *slog.Logger
	totalRequests   atomic.Int64
	totalErrors     atomic.Int64
	totalToolCalls  atomic.Int64
	firstChunkSum   atomic.Int64
	firstChunkCount atomic.Int64

	mu           sync.Mutex
	latencyBuckets []int64
}

func NewMetrics(logger *slog.Logger) *Metrics {
	return &Metrics{logger: logger}
}

func (m *Metrics) RecordRequest()                { m.totalRequests.Add(1) }
func (m *Metrics) RecordError()                  { m.totalErrors.Add(1) }
func (m *Metrics) RecordToolCalls(n int)         { m.totalToolCalls.Add(int64(n)) }

func (m *Metrics) RecordFirstChunkLatency(d time.Duration) {
	ms := d.Milliseconds()
	m.firstChunkSum.Add(ms)
	m.firstChunkCount.Add(1)
	m.mu.Lock()
	m.latencyBuckets = append(m.latencyBuckets, ms)
	m.mu.Unlock()
}

// Snapshot returns current metrics for logging/export.
type Snapshot struct {
	TotalRequests  int64   `json:"total_requests"`
	TotalErrors    int64   `json:"total_errors"`
	TotalToolCalls int64   `json:"total_tool_calls"`
	AvgFirstChunk  float64 `json:"avg_first_chunk_ms"`
	P99FirstChunk  int64   `json:"p99_first_chunk_ms"`
}

func (m *Metrics) Snapshot() Snapshot {
	total := m.totalRequests.Load()
	errors := m.totalErrors.Load()
	tools := m.totalToolCalls.Load()
	sum := m.firstChunkSum.Load()
	count := m.firstChunkCount.Load()

	var avg float64
	if count > 0 {
		avg = float64(sum) / float64(count)
	}

	m.mu.Lock()
	var p99 int64
	if len(m.latencyBuckets) > 0 {
		sorted := make([]int64, len(m.latencyBuckets))
		copy(sorted, m.latencyBuckets)
		sortInt64(sorted)
		idx := int(float64(len(sorted)) * 0.99)
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		p99 = sorted[idx]
	}
	m.mu.Unlock()

	return Snapshot{
		TotalRequests:  total,
		TotalErrors:    errors,
		TotalToolCalls: tools,
		AvgFirstChunk:  avg,
		P99FirstChunk:  p99,
	}
}

func (m *Metrics) LogPeriodic(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			snap := m.Snapshot()
			m.logger.Info("metrics_snapshot",
				"total_requests", snap.TotalRequests,
				"total_errors", snap.TotalErrors,
				"total_tool_calls", snap.TotalToolCalls,
				"avg_first_chunk_ms", snap.AvgFirstChunk,
				"p99_first_chunk_ms", snap.P99FirstChunk,
			)
		case <-stopCh:
			return
		}
	}
}

func sortInt64(a []int64) {
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}

// RenderPrometheus exports current metrics in Prometheus text exposition format.
func (m *Metrics) RenderPrometheus() string {
	snap := m.Snapshot()
	var b strings.Builder
	b.WriteString("# HELP xiaozhi_requests_total Total service requests.\n")
	b.WriteString("# TYPE xiaozhi_requests_total counter\n")
	b.WriteString(fmt.Sprintf("xiaozhi_requests_total %d\n", snap.TotalRequests))
	b.WriteString("# HELP xiaozhi_errors_total Total service errors.\n")
	b.WriteString("# TYPE xiaozhi_errors_total counter\n")
	b.WriteString(fmt.Sprintf("xiaozhi_errors_total %d\n", snap.TotalErrors))
	b.WriteString("# HELP xiaozhi_tool_calls_total Total detected tool calls.\n")
	b.WriteString("# TYPE xiaozhi_tool_calls_total counter\n")
	b.WriteString(fmt.Sprintf("xiaozhi_tool_calls_total %d\n", snap.TotalToolCalls))
	b.WriteString("# HELP xiaozhi_first_chunk_avg_ms Average first chunk latency in ms.\n")
	b.WriteString("# TYPE xiaozhi_first_chunk_avg_ms gauge\n")
	b.WriteString(fmt.Sprintf("xiaozhi_first_chunk_avg_ms %.2f\n", snap.AvgFirstChunk))
	b.WriteString("# HELP xiaozhi_first_chunk_p99_ms P99 first chunk latency in ms.\n")
	b.WriteString("# TYPE xiaozhi_first_chunk_p99_ms gauge\n")
	b.WriteString(fmt.Sprintf("xiaozhi_first_chunk_p99_ms %d\n", snap.P99FirstChunk))
	return b.String()
}
