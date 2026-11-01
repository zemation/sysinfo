//go:build windows

package system

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	procGetLogicalDrives   = modKernel32.NewProc("GetLogicalDrives")
	procGetDriveTypeW      = modKernel32.NewProc("GetDriveTypeW")
	procGetDiskFreeSpaceEx = modKernel32.NewProc("GetDiskFreeSpaceExW")
)

const (
	driveRemovable = 2
	driveFixed     = 3
)

func getDiskSpace(path string) (freeBytes, totalBytes uint64, err error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}
	var freeAvail, total, totalFree uint64
	ret, _, callErr := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeAvail)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if ret == 0 {
		return 0, 0, callErr
	}
	return totalFree, total, nil
}

func GetDiskMounts() ([]DiskInfo, error) {
	mask, _, _ := procGetLogicalDrives.Call()
	if mask == 0 {
		return nil, fmt.Errorf("failed to list logical drives")
	}

	var disks []DiskInfo
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		drive := fmt.Sprintf("%c:\\", 'A'+i)
		drivePtr, err := syscall.UTF16PtrFromString(drive)
		if err != nil {
			continue
		}

		driveType, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(drivePtr)))
		if driveType != driveFixed && driveType != driveRemovable {
			continue
		}

		freeBytes, totalBytes, err := getDiskSpace(drive)
		if err != nil || totalBytes == 0 {
			continue
		}

		total := float64(totalBytes) / 1024 / 1024 / 1024
		free := float64(freeBytes) / 1024 / 1024 / 1024
		used := total - free
		percent := (used / total) * 100

		disks = append(disks, DiskInfo{
			Mount:   drive,
			Total:   fmt.Sprintf("%.1f GB", total),
			Used:    fmt.Sprintf("%.1f GB", used),
			Free:    fmt.Sprintf("%.1f GB", free),
			Percent: fmt.Sprintf("%.0f%%", percent),
		})
	}

	return disks, nil
}
