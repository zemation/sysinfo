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

type procSocketOwner struct {
	pid     string
	command string
}

func GetListeningPorts() ([]PortInfo, error) {
	inodeMap := buildSocketInodeMap()
	seen := make(map[string]bool)
	var ports []PortInfo

	sources := []struct {
		path  string
		proto string
	}{
		{"/proc/net/tcp", "TCP"},
		{"/proc/net/tcp6", "TCP"},
		{"/proc/net/udp", "UDP"},
		{"/proc/net/udp6", "UDP"},
	}

	for _, src := range sources {
		parsed, err := parseProcNet(src.path, src.proto, inodeMap, seen)
		if err == nil {
			ports = append(ports, parsed...)
		}
	}

	return ports, nil
}

func parseProcNet(path, proto string, inodeMap map[string]procSocketOwner, seen map[string]bool) ([]PortInfo, error) {
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
		// TCP listen is 0A (TCP_LISTEN); unconnected bound UDP sockets use 07 (TCP_CLOSE) or 0A
		if proto == "TCP" && state != "0A" {
			continue
		}
		if proto == "UDP" && state != "07" && state != "0A" {
			continue
		}

		localAddr := fields[1]
		addrParts := strings.Split(localAddr, ":")
		if len(addrParts) < 2 {
			continue
		}
		portHex := addrParts[len(addrParts)-1]
		portNum, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			continue
		}

		portStr := strconv.Itoa(int(portNum))
		key := fmt.Sprintf("%s:%s", proto, portStr)
		if seen[key] {
			continue
		}
		seen[key] = true

		inode := fields[9]
		owner, ok := inodeMap[inode]
		if !ok {
			owner = procSocketOwner{pid: "?", command: "?"}
		}

		ports = append(ports, PortInfo{
			Port:    portStr,
			Proto:   proto,
			State:   "LISTEN",
			PID:     owner.pid,
			Command: owner.command,
		})
	}

	return ports, nil
}

func buildSocketInodeMap() map[string]procSocketOwner {
	inodeMap := make(map[string]procSocketOwner)
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return inodeMap
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		fdPath := fmt.Sprintf("/proc/%s/fd", pid)
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		var comm string
		commLoaded := false

		for _, fd := range fds {
			link, err := os.Readlink(fmt.Sprintf("%s/%s", fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if strings.HasPrefix(link, "socket:[") && strings.HasSuffix(link, "]") {
				inode := link[8 : len(link)-1]
				if !commLoaded {
					commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%s/comm", pid))
					if err != nil {
						comm = "?"
					} else {
						comm = strings.TrimSpace(string(commBytes))
					}
					commLoaded = true
				}
				inodeMap[inode] = procSocketOwner{pid: pid, command: comm}
			}
		}
	}
	return inodeMap
}
