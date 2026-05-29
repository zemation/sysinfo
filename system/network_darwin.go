//go:build darwin

package system

import (
	"fmt"
	"net"
	"os/exec"
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
	out, err := exec.Command("netstat", "-I", name, "-b").Output()
	if err != nil {
		return "0 B", "0 B"
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return "0 B", "0 B"
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 10 {
		return "0 B", "0 B"
	}

	rxBytes, _ := strconv.ParseFloat(fields[6], 64)
	txBytes, _ := strconv.ParseFloat(fields[9], 64)

	return formatBytes(rxBytes), formatBytes(txBytes)
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

	tcpPorts, err := parseLsof("TCP")
	if err == nil {
		ports = append(ports, tcpPorts...)
	}

	udpPorts, err := parseLsof("UDP")
	if err == nil {
		ports = append(ports, udpPorts...)
	}

	return ports, nil
}

func parseLsof(proto string) ([]PortInfo, error) {
	var args []string
	if proto == "TCP" {
		args = []string{"-nP", "-iTCP", "-sTCP:LISTEN"}
	} else {
		args = []string{"-nP", "-iUDP"}
	}

	out, err := exec.Command("lsof", args...).Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var ports []PortInfo

	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		command := fields[0]
		pid := fields[1]
		name := fields[len(fields)-1]

		idx := strings.LastIndex(name, ":")
		if idx == -1 {
			continue
		}
		portStr := name[idx+1:]
		if _, err := strconv.Atoi(portStr); err != nil {
			continue
		}

		key := fmt.Sprintf("%s:%s", proto, portStr)
		if seen[key] {
			continue
		}
		seen[key] = true

		ports = append(ports, PortInfo{
			Port:    portStr,
			Proto:   proto,
			State:   "LISTEN",
			PID:     pid,
			Command: command,
		})
	}

	return ports, nil
}