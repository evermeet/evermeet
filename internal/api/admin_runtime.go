package api

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// adminRuntimeSnapshot returns OS/arch, Go runtime build, process memory stats,
// CPU topology, and host load averages when available (e.g. /proc/loadavg on Linux).
func adminRuntimeSnapshot() map[string]any {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	mem := map[string]any{
		"heap_alloc_bytes":  ms.HeapAlloc,
		"heap_inuse_bytes":  ms.HeapInuse,
		"heap_sys_bytes":    ms.HeapSys,
		"stack_inuse_bytes": ms.StackInuse,
		"sys_bytes":         ms.Sys,
		"gc_count":          ms.NumGC,
		"gc_cpu_fraction":   ms.GCCPUFraction,
	}
	cpu := map[string]any{
		"num_cpu":    runtime.NumCPU(),
		"gomaxprocs": runtime.GOMAXPROCS(0),
		"goroutines": runtime.NumGoroutine(),
	}
	if l1, l5, l15, ok := hostLoadAverage(); ok {
		cpu["loadavg_1"] = l1
		cpu["loadavg_5"] = l5
		cpu["loadavg_15"] = l15
	}
	return map[string]any{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"memory":     mem,
		"cpu":        cpu,
	}
}

func hostLoadAverage() (one, five, fifteen float64, ok bool) {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, false
	}
	fields := strings.Fields(string(raw))
	if len(fields) < 3 {
		return 0, 0, 0, false
	}
	one, err1 := strconv.ParseFloat(fields[0], 64)
	five, err2 := strconv.ParseFloat(fields[1], 64)
	fifteen, err3 := strconv.ParseFloat(fields[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return one, five, fifteen, true
}
