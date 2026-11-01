//go:build windows

package system

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	modKernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64       = modKernel32.NewProc("GetTickCount64")
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func GetOSInfo() (distro, kernel string) {
	distro = "Windows"
	kernel = "unknown"

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return distro, kernel
	}
	defer k.Close()

	productName, _, _ := k.GetStringValue("ProductName")
	displayVersion, _, _ := k.GetStringValue("DisplayVersion")
	buildNumber, _, _ := k.GetStringValue("CurrentBuildNumber")
	ubr, _, ubrErr := k.GetIntegerValue("UBR")

	// Windows 11 builds (>= 22000) still report "Windows 10" in ProductName in registry
	if buildNumber >= "22000" && strings.Contains(productName, "Windows 10") {
		productName = strings.Replace(productName, "Windows 10", "Windows 11", 1)
	}

	if productName != "" {
		if displayVersion != "" {
			distro = fmt.Sprintf("%s (%s)", productName, displayVersion)
		} else {
			distro = productName
		}
	}

	if buildNumber != "" {
		if ubrErr == nil {
			kernel = fmt.Sprintf("10.0.%s.%d", buildNumber, ubr)
		} else {
			kernel = fmt.Sprintf("10.0.%s", buildNumber)
		}
	}

	return distro, kernel
}

func GetMemory() (total, used string) {
	var mem memoryStatusEx
	mem.dwLength = uint32(unsafe.Sizeof(mem))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	if ret == 0 {
		return "unknown", "unknown"
	}

	totalGB := float64(mem.ullTotalPhys) / 1024 / 1024 / 1024
	usedGB := float64(mem.ullTotalPhys-mem.ullAvailPhys) / 1024 / 1024 / 1024
	return fmt.Sprintf("%.1f GB", totalGB), fmt.Sprintf("%.1f GB", usedGB)
}

func GetCPU() string {
	cores := runtime.NumCPU()
	model := "unknown"

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\Description\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		if val, _, err := k.GetStringValue("ProcessorNameString"); err == nil && strings.TrimSpace(val) != "" {
			model = strings.TrimSpace(val)
		}
	}

	return fmt.Sprintf("%s (%d cores)", model, cores)
}

func GetDisk() string {
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}
	if !strings.HasSuffix(drive, `\`) {
		drive += `\`
	}

	freeBytes, totalBytes, err := getDiskSpace(drive)
	if err != nil {
		return "unknown"
	}

	total := float64(totalBytes) / 1024 / 1024 / 1024
	free := float64(freeBytes) / 1024 / 1024 / 1024
	used := total - free
	return fmt.Sprintf("%.1f GB / %.1f GB", used, total)
}

func GetUptime() string {
	ret, _, _ := procGetTickCount64.Call()
	if ret == 0 {
		return "unknown"
	}

	uptimeSeconds := uint64(ret) / 1000
	days := uptimeSeconds / 86400
	hours := (uptimeSeconds % 86400) / 3600
	minutes := (uptimeSeconds % 3600) / 60
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

func GetLoadAverage() string {
	return "n/a (Windows)"
}
