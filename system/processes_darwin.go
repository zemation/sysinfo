//go:build darwin

package system

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func getProcessList() ([]Process, error) {
	out, err := exec.Command("ps", "axo", "pid=,%cpu=,%mem=,comm=").Output()
	if err != nil {
		return nil, err
	}

	var processes []Process
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
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
		name := strings.Join(fields[3:], " ")
		name = filepath.Base(name)

		processes = append(processes, Process{
			PID:     pid,
			Command: name,
			CPU:     fmt.Sprintf("%s%%", cpu),
			Memory:  fmt.Sprintf("%s%%", mem),
		})
	}

	return processes, nil
}
