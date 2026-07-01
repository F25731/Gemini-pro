package app

import (
	"sync"
	"time"
)

type MetricSnapshot struct {
	Total        int64   `json:"total"`
	Success     int64   `json:"success"`
	Failed      int64   `json:"failed"`
	SuccessRate float64 `json:"successRate"`
	AvgLatencyMs int64  `json:"avgLatencyMs"`
	LastLatencyMs int64  `json:"lastLatencyMs"`
}

type Metrics struct {
	mu     sync.RWMutex
	groups map[string]*metricCounter
}

type metricCounter struct {
	total         int64
	success       int64
	failed        int64
	totalLatency  int64
	lastLatencyMs int64
}

func NewMetrics() *Metrics {
	return &Metrics{groups: map[string]*metricCounter{}}
}

func (m *Metrics) Record(group string, ok bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	counter := m.groups[group]
	if counter == nil {
		counter = &metricCounter{}
		m.groups[group] = counter
	}
	latencyMs := latency.Milliseconds()
	counter.total++
	counter.totalLatency += latencyMs
	counter.lastLatencyMs = latencyMs
	if ok {
		counter.success++
	} else {
		counter.failed++
	}
}

func (m *Metrics) Snapshot() map[string]MetricSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]MetricSnapshot, len(m.groups))
	for key, counter := range m.groups {
		out[key] = snapshotCounter(counter)
	}
	return out
}

func snapshotCounter(counter *metricCounter) MetricSnapshot {
	s := MetricSnapshot{
		Total:         counter.total,
		Success:       counter.success,
		Failed:        counter.failed,
		LastLatencyMs: counter.lastLatencyMs,
	}
	if counter.total > 0 {
		s.SuccessRate = float64(counter.success) / float64(counter.total)
		s.AvgLatencyMs = counter.totalLatency / counter.total
	}
	return s
}
