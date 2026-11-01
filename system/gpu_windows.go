//go:build windows

package system

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func GetGPUInfo() GPUInfo {
	// Try NVIDIA first via nvidia-smi
	if gpu, ok := getNvidiaGPU(); ok {
		return gpu
	}

	// Fallback to Windows Display Adapter class registry
	if gpu, ok := getRegistryGPU(); ok {
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

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return GPUInfo{}, false
	}

	fields := strings.Split(strings.TrimSpace(lines[0]), ",")
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

func getRegistryGPU() (GPUInfo, bool) {
	const displayClassKey = `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, displayClassKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return GPUInfo{}, false
	}
	defer k.Close()

	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return GPUInfo{}, false
	}

	for _, sub := range subkeys {
		if len(sub) != 4 || sub < "0000" || sub > "9999" {
			continue
		}
		subKey, err := registry.OpenKey(registry.LOCAL_MACHINE, displayClassKey+`\`+sub, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		desc, _, err := subKey.GetStringValue("DriverDesc")
		driverVer, _, _ := subKey.GetStringValue("DriverVersion")
		subKey.Close()

		if err == nil && strings.TrimSpace(desc) != "" {
			if driverVer == "" {
				driverVer = "n/a"
			}
			return GPUInfo{
				Name:          strings.TrimSpace(desc),
				DriverVersion: fmt.Sprintf("%s", strings.TrimSpace(driverVer)),
				MemoryTotal:   "n/a",
				MemoryUsed:    "n/a",
				Utilization:   "n/a",
			}, true
		}
	}

	return GPUInfo{}, false
}
