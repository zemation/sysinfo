//go:build darwin

package system

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func GetIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

func GetOSInfo() (distro, kernel string) {
	name := swvers("ProductName")
	version := swvers("ProductVersion")
	if name == "" {
		distro = "macOS"
	} else {
		distro = name + " " + version
	}

	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		kernel = "unknown"
	} else {
		kernel = strings.TrimSpace(string(out))
	}

	return distro, kernel
}

func swvers(key string) string {
	out, err := exec.Command("sw_vers", "-"+key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetMemory() (total, used string) {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return "unknown", "unknown"
	}
	totalBytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return "unknown", "unknown"
	}

	freeBytes := vmStatFreeBytes()
	usedBytes := totalBytes - freeBytes

	totalGB := totalBytes / 1024 / 1024 / 1024
	usedGB := usedBytes / 1024 / 1024 / 1024
	return fmt.Sprintf("%.1f GB", totalGB), fmt.Sprintf("%.1f GB", usedGB)
}

func vmStatFreeBytes() float64 {
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0
	}

	pageSize := float64(4096)
	pOut, err := exec.Command("sysctl", "-n", "hw.pagesize").Output()
	if err == nil {
		if v, err := strconv.ParseFloat(strings.TrimSpace(string(pOut)), 64); err == nil {
			pageSize = v
		}
	}

	var freePages float64
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Pages free:") || strings.HasPrefix(line, "Pages inactive:") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				last := strings.TrimSuffix(parts[len(parts)-1], ".")
				v, _ := strconv.ParseFloat(last, 64)
				freePages += v
			}
		}
	}
	return freePages * pageSize
}

func GetCPU() string {
	model := sysctlString("machdep.cpu.brand_string")
	if model == "" {
		model = sysctlString("hw.model")
	}
	cores := sysctlString("hw.logicalcpu")
	if cores != "" {
		return fmt.Sprintf("%s (%s cores)", model, cores)
	}
	return model
}

func GetDisk() string {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return "unknown"
	}

	total := float64(stat.Blocks) * float64(stat.Bsize) / 1024 / 1024 / 1024
	free := float64(stat.Bfree) * float64(stat.Bsize) / 1024 / 1024 / 1024
	used := total - free
	return fmt.Sprintf("%.1f GB / %.1f GB", used, total)
}

func GetUptime() string {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return "unknown"
	}

	// Output looks like: { sec = 1748000000, usec = 0 } ...
	s := string(out)
	secIdx := strings.Index(s, "sec = ")
	if secIdx == -1 {
		return "unknown"
	}
	s = s[secIdx+6:]
	end := strings.IndexAny(s, ", }")
	if end == -1 {
		return "unknown"
	}
	bootSec, err := strconv.ParseInt(strings.TrimSpace(s[:end]), 10, 64)
	if err != nil {
		return "unknown"
	}

	var now syscall.Timeval
	syscall.Gettimeofday(&now)
	uptimeSecs := now.Sec - bootSec

	days := uptimeSecs / 86400
	hours := (uptimeSecs % 86400) / 3600
	minutes := (uptimeSecs % 3600) / 60
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

func GetLoadAverage() string {
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		return "unknown"
	}
	// Output looks like: { 0.12 0.15 0.10 }
	s := strings.Trim(strings.TrimSpace(string(out)), "{}")
	fields := strings.Fields(s)
	if len(fields) < 3 {
		return "unknown"
	}
	return fmt.Sprintf("%s, %s, %s (1, 5, 15 min)", fields[0], fields[1], fields[2])
}

func GetArch() string {
	return runtime.GOARCH
}

func sysctlString(name string) string {
	out, err := exec.Command("sysctl", "-n", name).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}