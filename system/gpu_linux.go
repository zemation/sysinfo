//go:build linux

package system

import (
	"os"
	"os/exec"
	"strings"
)

func GetGPUInfo() GPUInfo {
	// Try NVIDIA first
	if gpu, ok := getNvidiaGPU(); ok {
		return gpu
	}

	// Try AMD
	if gpu, ok := getAMDGPU(); ok {
		return gpu
	}

	// Try Intel via sysfs
	if gpu, ok := getIntelGPU(); ok {
		return gpu
	}

	return GPUInfo{Name: "No GPU detected"}
}

func getNvidiaGPU() (GPUInfo, bool) {
	out, err := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,driver_version,memory.total,memory.used,utilization.gpu",
		"--format=csv,noheader",
	).Output()
	if err != nil {
		return GPUInfo{}, false
	}

	fields := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(fields) < 5 {
		return GPUInfo{}, false
	}

	return GPUInfo{
		Name:          strings.TrimSpace(fields[0]),
		DriverVersion: strings.TrimSpace(fields[1]),
		MemoryTotal:   strings.TrimSpace(fields[2]),
		MemoryUsed:    strings.TrimSpace(fields[3]),
		Utilization:   strings.TrimSpace(fields[4]),
	}, true
}

func getAMDGPU() (GPUInfo, bool) {
	out, err := exec.Command("rocm-smi", "--showproductname", "--showmeminfo", "vram", "--showuse", "--csv").Output()
	if err != nil {
		return GPUInfo{}, false
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return GPUInfo{}, false
	}

	// rocm-smi csv: card,name,vram_total,vram_used,gpu_use
	fields := strings.Split(lines[1], ",")
	if len(fields) < 5 {
		return GPUInfo{Name: "AMD GPU", DriverVersion: "rocm-smi"}, true
	}

	return GPUInfo{
		Name:          strings.TrimSpace(fields[1]),
		DriverVersion: "ROCm",
		MemoryTotal:   strings.TrimSpace(fields[2]),
		MemoryUsed:    strings.TrimSpace(fields[3]),
		Utilization:   strings.TrimSpace(fields[4]),
	}, true
}

func getIntelGPU() (GPUInfo, bool) {
	// Check for Intel GPU via sysfs drm entries
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return GPUInfo{}, false
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "card") || strings.Contains(entry.Name(), "-") {
			continue
		}
		vendorPath := "/sys/class/drm/" + entry.Name() + "/device/vendor"
		data, err := os.ReadFile(vendorPath)
		if err != nil {
			continue
		}
		vendor := strings.TrimSpace(string(data))
		if vendor == "0x8086" { // Intel vendor ID
			namePath := "/sys/class/drm/" + entry.Name() + "/device/uevent"
			nameData, _ := os.ReadFile(namePath)
			name := "Intel Integrated GPU"
			for _, line := range strings.Split(string(nameData), "\n") {
				if strings.HasPrefix(line, "PCI_ID=") {
					name = "Intel GPU (" + strings.TrimPrefix(line, "PCI_ID=") + ")"
					break
				}
			}
			return GPUInfo{
				Name:          name,
				DriverVersion: "i915",
				MemoryTotal:   "shared",
				MemoryUsed:    "n/a",
				Utilization:   "n/a",
			}, true
		}
	}

	return GPUInfo{}, false
}
