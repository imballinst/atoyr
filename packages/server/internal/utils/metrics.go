package utils

import (
	"runtime"
	"slices"
	"time"
)

func GetTimePercentile(values []time.Duration, percentile int) int64 {
	if len(values) == 0 {
		return 0
	}

	// Sort a copy
	sorted := make([]time.Duration, len(values))
	copy(sorted, values)
	slices.Sort(sorted)

	index := (percentile * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index].Milliseconds()
}

func GetMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024 // Convert to MB
}
