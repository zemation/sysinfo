//go:build linux

package system

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func getProcessList() ([]Process, error) {
	var processes []Process

	memTotal := getTotalMemoryKB()
	totalCPU := getTotalCPUTime()

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid := entry.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		proc, err := readProcess(pid, memTotal, totalCPU)
		if err != nil {
			continue
		}
		processes = append(processes, proc)
	}

	return processes, nil
}

func readProcess(pid string, memTotal float64, totalCPU float64) (Process, error) {
	commPath := filepath.Join("/proc", pid, "comm")
	commBytes, err := os.ReadFile(commPath)
	if err != nil {
		return Process{}, err
	}
	command := strings.TrimSpace(string(commBytes))

	statPath := filepath.Join("/proc", pid, "stat")
	statBytes, err := os.ReadFile(statPath)
	if err != nil {
		return Process{}, err
	}

	fields := strings.Fields(string(statBytes))
	if len(fields) < 24 {
		return Process{}, fmt.Errorf("unexpected stat format")
	}

	utime, _ := strconv.ParseFloat(fields[13], 64)
	stime, _ := strconv.ParseFloat(fields[14], 64)
	processCPU := utime + stime

	rss, _ := strconv.ParseFloat(fields[23], 64)
	memBytes := rss * 4096
	memPercent := (memBytes / 1024 / (memTotal)) * 100

	cpuPercent := 0.0
	if totalCPU > 0 {
		cpuPercent = (processCPU / totalCPU) * 100
	}

	return Process{
		PID:     pid,
		Command: command,
		CPU:     fmt.Sprintf("%.1f%%", cpuPercent),
		Memory:  fmt.Sprintf("%.1f%%", memPercent),
	}, nil
}

func getTotalMemoryKB() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 1
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, _ := strconv.ParseFloat(fields[1], 64)
				return val
			}
		}
	}
	return 1
}

func getTotalCPUTime() float64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 1
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			var total float64
			for _, f := range fields[1:] {
				val, _ := strconv.ParseFloat(f, 64)
				total += val
			}
			return total
		}
	}
	return 1
}

func GetTopProcessesByCPU(n int) ([]Process, error) {
	procs, err := getProcessList()
	if err != nil {
		return nil, err
	}

	sort.Slice(procs, func(i, j int) bool {
		iVal, _ := strconv.ParseFloat(strings.TrimSuffix(procs[i].CPU, "%"), 64)
		jVal, _ := strconv.ParseFloat(strings.TrimSuffix(procs[j].CPU, "%"), 64)
		return iVal > jVal
	})

	if len(procs) > n {
		procs = procs[:n]
	}
	return procs, nil
}

func GetTopProcessesByMemory(n int) ([]Process, error) {
	procs, err := getProcessList()
	if err != nil {
		return nil, err
	}

	sort.Slice(procs, func(i, j int) bool {
		iVal, _ := strconv.ParseFloat(strings.TrimSuffix(procs[i].Memory, "%"), 64)
		jVal, _ := strconv.ParseFloat(strings.TrimSuffix(procs[j].Memory, "%"), 64)
		return iVal > jVal
	})

	if len(procs) > n {
		procs = procs[:n]
	}
	return procs, nil
}