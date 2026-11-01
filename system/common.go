package system

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
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

func GetArch() string {
	return runtime.GOARCH
}

func formatBytes(bytes float64) string {
	if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", bytes/1024/1024/1024)
	} else if bytes >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", bytes/1024/1024)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.1f KB", bytes/1024)
	}
	return fmt.Sprintf("%.0f B", bytes)
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
