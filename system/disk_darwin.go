//go:build darwin

package system

import (
	"fmt"
	"syscall"
)

func GetDiskMounts() ([]DiskInfo, error) {
	var buf [512]syscall.Statfs_t
	n, err := syscall.Getfsstat(buf[:], 1) // MNT_WAIT = 1
	if err != nil {
		return nil, err
	}

	var disks []DiskInfo
	seen := make(map[string]bool)

	for i := 0; i < n; i++ {
		fs := buf[i]
		fstype := int8SliceToString(fs.Fstypename[:])

		switch fstype {
		case "devfs", "autofs", "nullfs", "fdesc", "volfs":
			continue
		}

		mount := int8SliceToString(fs.Mntonname[:])
		if seen[mount] {
			continue
		}
		seen[mount] = true

		bsize := float64(fs.Bsize)
		total := float64(fs.Blocks) * bsize / 1024 / 1024 / 1024
		free := float64(fs.Bfree) * bsize / 1024 / 1024 / 1024
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

func int8SliceToString(b []int8) string {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return string(buf)
}