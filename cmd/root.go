package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var rootCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "A lightweight system information tool",
	Run: func(cmd *cobra.Command, args []string) {
		distro, kernel := system.GetOSInfo()
		memTotal, memUsed := system.GetMemory()
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
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}