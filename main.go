package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"net"
)

func getOSInfo() (distro, kernel string) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		distro = "unknown"
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				distro = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}

	file2, err := os.Open("/proc/version")
	if err != nil {
		kernel = "unknown"
	} else {
		defer file2.Close()
		scanner := bufio.NewScanner(file2)
		if scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 3 {
				kernel = fields[2]
			}
		}
	}
	return distro, kernel
}

func getMemory() (total, used string) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "unknown", "unknown"
	}
	defer file.Close()

	var memTotal, memAvailable float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value := fields[1]
		switch fields[0] {
		case "MemTotal:":
			fmt.Sscanf(value, "%f", &memTotal)
		case "MemAvailable:":
			fmt.Sscanf(value, "%f", &memAvailable)
		}
	}

	totalGB := memTotal / 1024 / 1024
	usedGB := (memTotal - memAvailable) / 1024 / 1024
	return fmt.Sprintf("%.1f GB", totalGB), fmt.Sprintf("%.1f GB", usedGB)
}

func getCPU() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	var cores int
	var model string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if strings.HasPrefix(line, "processor") {
			cores++
		}
		if strings.HasPrefix(line, "model name") {
			model = strings.Join(fields[3:], " ")
		}
	}
	return fmt.Sprintf("%s (%d cores)", model, cores)
}

func getDisk() string {
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

func getUptime() string {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	var uptimeSeconds float64
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
	}

	days := int(uptimeSeconds / 86400)
	hours := int(uptimeSeconds/3600) % 24
	minutes := int(uptimeSeconds/60) % 60
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

func getIPAddress() string {
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

func getLoadAverage() string {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			return fmt.Sprintf("%s, %s, %s (1, 5, 15 min)", fields[0], fields[1], fields[2])
		}
	}
	return "unknown"
}

func main() {
	hostname, _ := os.Hostname()
	distro, kernel := getOSInfo()
	memTotal, memUsed := getMemory()
	fmt.Println("Host:   ", hostname)
	fmt.Println("IP:     ", getIPAddress())
	fmt.Println("OS:     ", distro)
	fmt.Println("Kernel: ", kernel)
	fmt.Println("Arch:   ", runtime.GOARCH)
	fmt.Println("CPU:    ", getCPU())
	fmt.Println("Memory: ", memUsed, "/", memTotal)
	fmt.Println("Disk:   ", getDisk())
	fmt.Println("Uptime: ", getUptime())
        fmt.Println("Load:   ", getLoadAverage())
}
