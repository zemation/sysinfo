package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

type DiskInfo struct {
	Mount   string
	Total   string
	Used    string
	Free    string
	Percent string
}

func GetDiskMounts() ([]DiskInfo, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var disks []DiskInfo
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		device := fields[0]
		mount := fields[1]

		// skip non-real filesystems
		if !strings.HasPrefix(device, "/") {
			continue
		}

		// skip duplicates
		if seen[mount] {
			continue
		}
		seen[mount] = true

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err != nil {
			continue
		}

		total := float64(stat.Blocks) * float64(stat.Bsize) / 1024 / 1024 / 1024
		free := float64(stat.Bfree) * float64(stat.Bsize) / 1024 / 1024 / 1024
		used := total - free
		percent := 0.0
		if total > 0 {
			percent = (used / total) * 100
		}

		disks = append(disks, DiskInfo{
			Mount:   mount,
			Total:   fmt.Sprintf("%.1f GB", total),
			Used:    fmt.Sprintf("%.1f GB", used),
			Free:    fmt.Sprintf("%.1f GB", free),
			Percent: fmt.Sprintf("%.0f%%", percent),
		})
	}

	return disks, nil
}