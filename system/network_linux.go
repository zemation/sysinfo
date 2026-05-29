//go:build linux

package system

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func GetNetworkInterfaces() ([]NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []NetworkInterface

	for _, iface := range ifaces {
		ip := "n/a"
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
						break
					}
				}
			}
		}

		rx, tx := getInterfaceStats(iface.Name)

		result = append(result, NetworkInterface{
			Name: iface.Name,
			IP:   ip,
			RX:   rx,
			TX:   tx,
		})
	}

	return result, nil
}

func getInterfaceStats(name string) (rx, tx string) {
	rxPath := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", name)
	txPath := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", name)

	rxBytes := readStatFile(rxPath)
	txBytes := readStatFile(txPath)

	return formatBytes(rxBytes), formatBytes(txBytes)
}

func readStatFile(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	return val
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

func GetListeningPorts() ([]PortInfo, error) {
	var ports []PortInfo

	tcpPorts, err := parseProcNet("/proc/net/tcp", "TCP")
	if err == nil {
		ports = append(ports, tcpPorts...)
	}

	udpPorts, err := parseProcNet("/proc/net/udp", "UDP")
	if err == nil {
		ports = append(ports, udpPorts...)
	}

	return ports, nil
}

func parseProcNet(path, proto string) ([]PortInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []PortInfo
	scanner := bufio.NewScanner(file)
	scanner.Scan()

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		state := fields[3]
		if state != "0A" {
			continue
		}

		localAddr := fields[1]
		portHex := strings.Split(localAddr, ":")[1]
		portNum, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			continue
		}

		inode := fields[9]
		pid, command := findProcessByInode(inode)

		ports = append(ports, PortInfo{
			Port:    strconv.Itoa(int(portNum)),
			Proto:   proto,
			State:   "LISTEN",
			PID:     pid,
			Command: command,
		})
	}

	return ports, nil
}

func findProcessByInode(inode string) (pid, command string) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return "?", "?"
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}

		fdPath := fmt.Sprintf("/proc/%s/fd", entry.Name())
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(fmt.Sprintf("%s/%s", fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if strings.Contains(link, fmt.Sprintf("socket:[%s]", inode)) {
				commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%s/comm", entry.Name()))
				if err != nil {
					return entry.Name(), "?"
				}
				return entry.Name(), strings.TrimSpace(string(commBytes))
			}
		}
	}
	return "?", "?"
}