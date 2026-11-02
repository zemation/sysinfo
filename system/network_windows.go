//go:build windows

package system

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	modIphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfEntry2         = modIphlpapi.NewProc("GetIfEntry2")
	procConvertIdxLuid      = modIphlpapi.NewProc("ConvertInterfaceIndexToLuid")
	procGetExtendedTcpTable = modIphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modIphlpapi.NewProc("GetExtendedUdpTable")
)

const (
	afInet                 = 2
	afInet6                = 23
	tcpTableOwnerPidListen = 3 // TCP_TABLE_OWNER_PID_LISTENER
	udpTableOwnerPid       = 1 // UDP_TABLE_OWNER_PID
	mibTcpStateListen      = 2
)

type mibTcpRowOwnerPid struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

type mibTcp6RowOwnerPid struct {
	ucLocalAddr     [16]byte
	dwLocalScopeId  uint32
	dwLocalPort     uint32
	ucRemoteAddr    [16]byte
	dwRemoteScopeId uint32
	dwRemotePort    uint32
	dwState         uint32
	dwOwningPid     uint32
}

type mibUdpRowOwnerPid struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

type mibUdp6RowOwnerPid struct {
	ucLocalAddr    [16]byte
	dwLocalScopeId uint32
	dwLocalPort    uint32
	dwOwningPid    uint32
}

// mibIfRow2 represents MIB_IF_ROW2 from netioapi.h (1352 bytes on 64-bit Windows)
// InterfaceLuid is at offset 0 (uint64)
// InterfaceIndex is at offset 8 (uint32)
// InOctets is at offset 1224 (uint64)
// OutOctets is at offset 1280 (uint64)
type mibIfRow2 struct {
	data [1352]byte
}

func ntohs(port uint32) int {
	p := uint16(port)
	return int((p>>8)&0xff) | int((p&0xff)<<8)
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
	pidNames := getProcessNameMap()
	seen := make(map[string]bool)
	var ports []PortInfo

	addPort := func(proto string, port int, pid uint32) {
		portStr := strconv.Itoa(port)
		key := fmt.Sprintf("%s:%s", proto, portStr)
		if seen[key] {
			return
		}
		seen[key] = true

		pidStr := strconv.FormatUint(uint64(pid), 10)
		command := pidNames[pidStr]
		if command == "" {
			command = "?"
		}

		ports = append(ports, PortInfo{
			Port:    portStr,
			Proto:   proto,
			State:   "LISTEN",
			PID:     pidStr,
			Command: command,
		})
	}

	// TCP IPv4
	if buf, err := fetchIpHlpTable(procGetExtendedTcpTable, afInet, tcpTableOwnerPidListen); err == nil && len(buf) >= 4 {
		count := *(*uint32)(unsafe.Pointer(&buf[0]))
		rows := unsafe.Slice((*mibTcpRowOwnerPid)(unsafe.Pointer(&buf[4])), count)
		for _, r := range rows {
			if r.dwState == mibTcpStateListen {
				addPort("TCP", ntohs(r.dwLocalPort), r.dwOwningPid)
			}
		}
	}

	// TCP IPv6
	if buf, err := fetchIpHlpTable(procGetExtendedTcpTable, afInet6, tcpTableOwnerPidListen); err == nil && len(buf) >= 4 {
		count := *(*uint32)(unsafe.Pointer(&buf[0]))
		rows := unsafe.Slice((*mibTcp6RowOwnerPid)(unsafe.Pointer(&buf[4])), count)
		for _, r := range rows {
			if r.dwState == mibTcpStateListen {
				addPort("TCP", ntohs(r.dwLocalPort), r.dwOwningPid)
			}
		}
	}

	// UDP IPv4
	if buf, err := fetchIpHlpTable(procGetExtendedUdpTable, afInet, udpTableOwnerPid); err == nil && len(buf) >= 4 {
		count := *(*uint32)(unsafe.Pointer(&buf[0]))
		rows := unsafe.Slice((*mibUdpRowOwnerPid)(unsafe.Pointer(&buf[4])), count)
		for _, r := range rows {
			addPort("UDP", ntohs(r.dwLocalPort), r.dwOwningPid)
		}
	}

	// UDP IPv6
	if buf, err := fetchIpHlpTable(procGetExtendedUdpTable, afInet6, udpTableOwnerPid); err == nil && len(buf) >= 4 {
		count := *(*uint32)(unsafe.Pointer(&buf[0]))
		rows := unsafe.Slice((*mibUdp6RowOwnerPid)(unsafe.Pointer(&buf[4])), count)
		for _, r := range rows {
			addPort("UDP", ntohs(r.dwLocalPort), r.dwOwningPid)
		}
	}

	return ports, nil
}

func fetchIpHlpTable(proc *syscall.LazyProc, family uintptr, tableClass uintptr) ([]byte, error) {
	var size uint32
	proc.Call(0, uintptr(unsafe.Pointer(&size)), 1, family, tableClass, 0)
	if size == 0 {
		return nil, fmt.Errorf("empty table")
	}

	buf := make([]byte, size)
	ret, _, err := proc.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		1,
		family,
		tableClass,
		0,
	)
	if ret != 0 {
		return nil, err
	}
	return buf, nil
}
