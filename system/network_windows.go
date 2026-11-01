//go:build windows

package system

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modIphlpapi        = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfEntry2    = modIphlpapi.NewProc("GetIfEntry2")
	procConvertIdxLuid = modIphlpapi.NewProc("ConvertInterfaceIndexToLuid")
)

// mibIfRow2 represents MIB_IF_ROW2 from netioapi.h (1352 bytes on 64-bit Windows)
// InterfaceLuid is at offset 0 (uint64)
// InterfaceIndex is at offset 8 (uint32)
// InOctets is at offset 1224 (uint64)
// OutOctets is at offset 1280 (uint64)
type mibIfRow2 struct {
	data [1352]byte
}

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

		rx, tx := getInterfaceStats(iface.Index)

		result = append(result, NetworkInterface{
			Name: iface.Name,
			IP:   ip,
			RX:   rx,
			TX:   tx,
		})
	}

	return result, nil
}

func getInterfaceStats(ifIndex int) (rx, tx string) {
	var luid uint64
	ret, _, _ := procConvertIdxLuid.Call(uintptr(ifIndex), uintptr(unsafe.Pointer(&luid)))
	if ret != 0 {
		return "0 B", "0 B"
	}

	var row mibIfRow2
	*(*uint64)(unsafe.Pointer(&row.data[0])) = luid
	ret, _, _ = procGetIfEntry2.Call(uintptr(unsafe.Pointer(&row)))
	if ret != 0 {
		return "0 B", "0 B"
	}

	inOctets := *(*uint64)(unsafe.Pointer(&row.data[1224]))
	outOctets := *(*uint64)(unsafe.Pointer(&row.data[1280]))

	return formatBytes(float64(inOctets)), formatBytes(float64(outOctets))
}

func GetListeningPorts() ([]PortInfo, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, err
	}

	pidNames := getProcessNameMap()
	seen := make(map[string]bool)
	var ports []PortInfo

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		if proto != "TCP" && proto != "UDP" {
			continue
		}

		var state, pid string
		if proto == "TCP" {
			if len(fields) < 5 || fields[3] != "LISTENING" {
				continue
			}
			state = "LISTEN"
			pid = fields[4]
		} else {
			// UDP lines in netstat -ano: UDP <local> *:* <pid>
			state = "LISTEN"
			pid = fields[len(fields)-1]
		}

		localAddr := fields[1]
		idx := strings.LastIndex(localAddr, ":")
		if idx == -1 {
			continue
		}
		portStr := localAddr[idx+1:]
		if _, err := strconv.Atoi(portStr); err != nil {
			continue
		}

		key := fmt.Sprintf("%s:%s", proto, portStr)
		if seen[key] {
			continue
		}
		seen[key] = true

		command, ok := pidNames[pid]
		if !ok || command == "" {
			command = "?"
		}

		ports = append(ports, PortInfo{
			Port:    portStr,
			Proto:   proto,
			State:   state,
			PID:     pid,
			Command: command,
		})
	}

	return ports, nil
}
