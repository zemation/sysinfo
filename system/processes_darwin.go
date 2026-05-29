//go:build darwin

package system

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func getProcessList() ([]Process, error) {
	out, err := exec.Command("ps", "axo", "pid,%cpu,%mem,comm").Output()
	if err != nil {
		return nil, err
	}

	var processes []Process
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid := fields[0]
		cpu := fields[1]
		mem := fields[2]
		name := fields[3]

		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]

		processes = append(processes, Process{
			PID:     pid,
			Command: name,
			CPU:     fmt.Sprintf("%s%%", cpu),
			Memory:  fmt.Sprintf("%s%%", mem),
		})
	}

	return processes, nil
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