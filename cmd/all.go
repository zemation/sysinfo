package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

type AllInfo struct {
	System            SystemInfo                `json:"system"`
	DiskMounts        []system.DiskInfo         `json:"disk_mounts"`
	NetworkInterfaces []system.NetworkInterface `json:"network_interfaces"`
	TopCPUProcesses   []system.Process          `json:"top_cpu_processes"`
	TopMemProcesses   []system.Process          `json:"top_memory_processes"`
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Show all system information",
	Run: func(cmd *cobra.Command, args []string) {
		distro, kernel := system.GetOSInfo()
		memTotal, memUsed := system.GetMemory()
		gpu := system.GetGPUInfo()

		if jsonOutput {
			disks, _ := system.GetDiskMounts()
			ifaces, _ := system.GetNetworkInterfaces()
			cpuProcs, _ := system.GetTopProcessesByCPU(5)
			memProcs, _ := system.GetTopProcessesByMemory(5)

			info := AllInfo{
				System: SystemInfo{
					Host:     system.GetHostname(),
					IP:       system.GetIPAddress(),
					OS:       distro,
					Kernel:   kernel,
					Arch:     system.GetArch(),
					CPU:      system.GetCPU(),
					MemTotal: memTotal,
					MemUsed:  memUsed,
					Disk:     system.GetDisk(),
					Uptime:   system.GetUptime(),
					Load:     system.GetLoadAverage(),
					GPU:      gpu.Name,
				},
				DiskMounts:        disks,
				NetworkInterfaces: ifaces,
				TopCPUProcesses:   cpuProcs,
				TopMemProcesses:   memProcs,
			}
			out, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(out))
			return
		}

		// System overview
		fmt.Println("=== System ===")
		fmt.Println("Host:   ", system.GetHostname())
		fmt.Println("IP:     ", system.GetIPAddress())
		fmt.Println("OS:     ", distro)
		fmt.Println("Kernel: ", kernel)
		fmt.Println("Arch:   ", system.GetArch())
		fmt.Println("CPU:    ", system.GetCPU())
		fmt.Println("Memory: ", memUsed, "/", memTotal)
		fmt.Println("Disk:   ", system.GetDisk())
		fmt.Println("Uptime: ", system.GetUptime())
		fmt.Println("Load:   ", system.GetLoadAverage())
		fmt.Println("GPU:    ", gpu.Name)

		// Disk mounts
		fmt.Println("\n=== Disk Mounts ===")
		disks, err := system.GetDiskMounts()
		if err == nil {
			fmt.Printf("%-20s %-10s %-10s %-10s %s\n", "MOUNT", "TOTAL", "USED", "FREE", "USE%")
			fmt.Println("------------------------------------------------------")
			for _, d := range disks {
				fmt.Printf("%-20s %-10s %-10s %-10s %s\n", d.Mount, d.Total, d.Used, d.Free, system.Colorize(d.Percent, 70, 90))
			}
		}

		// Network interfaces
		fmt.Println("\n=== Network Interfaces ===")
		ifaces, err := system.GetNetworkInterfaces()
		if err == nil {
			fmt.Printf("%-12s %-18s %-12s %s\n", "INTERFACE", "IP", "RX", "TX")
			fmt.Println("------------------------------------------------------")
			for _, i := range ifaces {
				fmt.Printf("%-12s %-18s %-12s %s\n", i.Name, i.IP, i.RX, i.TX)
			}
		}

		// Top processes by CPU
		fmt.Println("\n=== Top Processes (CPU) ===")
		procs, err := system.GetTopProcessesByCPU(5)
		if err == nil {
			fmt.Printf("%-10s %-8s %s\n", "PID", "CPU%", "COMMAND")
			fmt.Println("-------------------------------")
			for _, p := range procs {
				fmt.Printf("%-10s %s %s\n", p.PID, system.ColorizePadded(p.CPU, 8, 70, 90), p.Command)
			}
		}

		// Top processes by memory
		fmt.Println("\n=== Top Processes (Memory) ===")
		memProcs, err := system.GetTopProcessesByMemory(5)
		if err == nil {
			fmt.Printf("%-10s %-8s %s\n", "PID", "MEM%", "COMMAND")
			fmt.Println("-------------------------------")
			for _, p := range memProcs {
				fmt.Printf("%-10s %s %s\n", p.PID, system.ColorizePadded(p.Memory, 8, 70, 90), p.Command)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
