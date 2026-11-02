package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	GPU      string `json:"gpu"`
}

var (
	jsonOutput    bool
	watchMode     bool
	watchInterval int
)

var rootCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "A lightweight system information tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !watchMode || cmd.Run == nil {
			return
		}
		origRun := cmd.Run
		cmd.Run = func(c *cobra.Command, a []string) {
			interval := time.Duration(watchInterval) * time.Second
			if interval < time.Second {
				interval = 2 * time.Second
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				if system.IsTerminal() && !jsonOutput {
					fmt.Print("\033[H\033[2J")
					fmt.Printf("Every %ds: %s (%s) — Press Ctrl+C to exit\n\n", int(interval.Seconds()), c.CommandPath(), time.Now().Format("15:04:05"))
				}
				origRun(c, a)

				select {
				case <-sigCh:
					return
				case <-ticker.C:
				}
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		distro, kernel := system.GetOSInfo()
		memTotal, memUsed := system.GetMemory()
		gpu := system.GetGPUInfo()

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
				GPU:      gpu.Name,
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
		fmt.Println("GPU:    ", gpu.Name)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&watchMode, "watch", "w", false, "Continuously refresh output")
	rootCmd.PersistentFlags().IntVarP(&watchInterval, "interval", "i", 2, "Refresh interval in seconds when using --watch")
}
