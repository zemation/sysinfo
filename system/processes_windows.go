//go:build windows

package system

import (
	"fmt"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	modPsapi                  = syscall.NewLazyDLL("psapi.dll")
	procGetProcessMemoryInfo  = modPsapi.NewProc("GetProcessMemoryInfo")
	procGetSystemTimes        = modKernel32.NewProc("GetSystemTimes")
	procOpenProcess           = modKernel32.NewProc("OpenProcess")
	procGetProcessTimes       = modKernel32.NewProc("GetProcessTimes")
	procCreateToolhelp32      = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW       = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW        = modKernel32.NewProc("Process32NextW")
)

const (
	th32csSnapProcess         = 0x00000002
	processQueryLimitedInfo   = 0x1000
	processVMRead             = 0x0010
)

type processEntry32W struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

type processMemoryCounters struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

func filetimeToUint64(ft syscall.Filetime) uint64 {
	return (uint64(ft.HighDateTime) << 32) | uint64(ft.LowDateTime)
}

func getProcessNameMap() map[string]string {
	names := make(map[string]string)
	snap, _, _ := procCreateToolhelp32.Call(th32csSnapProcess, 0)
	if snap == uintptr(syscall.InvalidHandle) || snap == 0 {
		return names
	}
	defer syscall.CloseHandle(syscall.Handle(snap))

	var entry processEntry32W
	entry.dwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		pid := strconv.FormatUint(uint64(entry.th32ProcessID), 10)
		names[pid] = syscall.UTF16ToString(entry.szExeFile[:])
		ret, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return names
}

func getProcessList() ([]Process, error) {
	snap, _, err := procCreateToolhelp32.Call(th32csSnapProcess, 0)
	if snap == uintptr(syscall.InvalidHandle) || snap == 0 {
		return nil, err
	}
	defer syscall.CloseHandle(syscall.Handle(snap))

	var mem memoryStatusEx
	mem.dwLength = uint32(unsafe.Sizeof(mem))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	totalPhys := float64(mem.ullTotalPhys)
	if totalPhys <= 0 {
		totalPhys = 1
	}

	retTick, _, _ := procGetTickCount64.Call()
	uptime100ns := float64(retTick) * 10000.0
	numCPU := float64(runtime.NumCPU())
	if numCPU < 1 {
		numCPU = 1
	}
	totalSystemCPU := uptime100ns * numCPU

	var processes []Process
	var entry processEntry32W
	entry.dwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		pid := entry.th32ProcessID
		name := syscall.UTF16ToString(entry.szExeFile[:])

		if pid != 0 {
			cpuPct, memPct := queryProcessMetrics(pid, totalPhys, totalSystemCPU)
			processes = append(processes, Process{
				PID:     strconv.FormatUint(uint64(pid), 10),
				Command: name,
				CPU:     fmt.Sprintf("%.1f%%", cpuPct),
				Memory:  fmt.Sprintf("%.1f%%", memPct),
			})
		}

		ret, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}

	return processes, nil
}

func queryProcessMetrics(pid uint32, totalPhys, totalSystemCPU float64) (cpuPct, memPct float64) {
	h, _, _ := procOpenProcess.Call(processQueryLimitedInfo|processVMRead, 0, uintptr(pid))
	if h == 0 {
		h, _, _ = procOpenProcess.Call(processQueryLimitedInfo, 0, uintptr(pid))
	}
	if h == 0 {
		return 0, 0
	}
	defer syscall.CloseHandle(syscall.Handle(h))

	var pmc processMemoryCounters
	pmc.cb = uint32(unsafe.Sizeof(pmc))
	retMem, _, _ := procGetProcessMemoryInfo.Call(h, uintptr(unsafe.Pointer(&pmc)), uintptr(pmc.cb))
	if retMem != 0 && totalPhys > 0 {
		memPct = (float64(pmc.WorkingSetSize) / totalPhys) * 100
	}

	var creation, exit, kernel, user syscall.Filetime
	retCPU, _, _ := procGetProcessTimes.Call(
		h,
		uintptr(unsafe.Pointer(&creation)),
		uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if retCPU != 0 && totalSystemCPU > 0 {
		procTime := float64(filetimeToUint64(kernel) + filetimeToUint64(user))
		cpuPct = (procTime / totalSystemCPU) * 100
	}

	return cpuPct, memPct
}
