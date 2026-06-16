//go:build darwin

package system

import (
	"os/exec"
	"strings"
)

func GetGPUInfo() GPUInfo {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return GPUInfo{Name: "No GPU detected"}
	}

	var name, vram string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chipset Model:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "Chipset Model:"))
		}
		if strings.HasPrefix(line, "VRAM") {
			// e.g. "VRAM (Total): 8 GB" or "VRAM (Dynamic, Max):  2 GB"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				vram = strings.TrimSpace(parts[1])
			}
		}
	}

	if name == "" {
		return GPUInfo{Name: "No GPU detected"}
	}

	return GPUInfo{
		Name:          name,
		DriverVersion: "n/a",
		MemoryTotal:   vram,
		MemoryUsed:    "n/a",
		Utilization:   "n/a",
	}
}
