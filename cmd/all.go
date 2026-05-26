package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Show all system information",
	Run: func(cmd *cobra.Command, args []string) {
		// System overview
		distro, kernel := system.GetOSInfo()
		memTotal, memUsed := system.GetMemory()
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

		// Disk mounts
		fmt.Println("\n=== Disk Mounts ===")
		disks, err := system.GetDiskMounts()
		if err == nil {
			fmt.Printf("%-20s %-10s %-10s %-10s %s\n", "MOUNT", "TOTAL", "USED", "FREE", "USE%")
			fmt.Println("------------------------------------------------------")
			for _, d := range disks {
				fmt.Printf("%-20s %-10s %-10s %-10s %s\n", d.Mount, d.Total, d.Used, d.Free, d.Percent)
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
				fmt.Printf("%-10s %-8s %s\n", p.PID, p.CPU, p.Command)
			}
		}

		// Top processes by memory
		fmt.Println("\n=== Top Processes (Memory) ===")
		memProcs, err := system.GetTopProcessesByMemory(5)
		if err == nil {
			fmt.Printf("%-10s %-8s %s\n", "PID", "MEM%", "COMMAND")
			fmt.Println("-------------------------------")
			for _, p := range memProcs {
				fmt.Printf("%-10s %-8s %s\n", p.PID, p.Memory, p.Command)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}