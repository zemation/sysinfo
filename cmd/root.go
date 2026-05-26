package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

type SystemInfo struct {
	Host     string `json:"host"`
	IP       string `json:"ip"`
	OS       string `json:"os"`
	Kernel   string `json:"kernel"`
	Arch     string `json:"arch"`
	CPU      string `json:"cpu"`
	MemTotal string `json:"mem_total"`
	MemUsed  string `json:"mem_used"`
	Disk     string `json:"disk"`
	Uptime   string `json:"uptime"`
	Load     string `json:"load"`
}

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "A lightweight system information tool",
	Run: func(cmd *cobra.Command, args []string) {
		distro, kernel := system.GetOSInfo()
		memTotal, memUsed := system.GetMemory()

		if jsonOutput {
			info := SystemInfo{
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
			}
			out, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(out))
			return
		}

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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
}