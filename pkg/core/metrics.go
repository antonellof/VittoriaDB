package core

import (
	"sync/atomic"
	"time"
)

// Process-wide search counters (single embedded server per process).
var (
	searchCount   atomic.Uint64
	searchNsSum   atomic.Uint64
)

// RecordSearchObserved increments global search statistics (duration wall-clock for one search).
func RecordSearchObserved(d time.Duration) {
	searchCount.Add(1)
	searchNsSum.Add(uint64(d.Nanoseconds()))
}

// SearchMetricsSnapshot returns cumulative search stats since process start.
func SearchMetricsSnapshot() (total uint64, avgLatency time.Duration) {
	n := searchCount.Load()
	if n == 0 {
		return 0, 0
	}
	ns := searchNsSum.Load()
	return n, time.Duration(ns / n)
}
