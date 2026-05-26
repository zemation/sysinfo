package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var processesCmd = &cobra.Command{
	Use:   "processes",
	Short: "Show top processes by resource usage",
}

var processesCPUCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Show top 5 processes by CPU usage",
	Run: func(cmd *cobra.Command, args []string) {
		procs, err := system.GetTopProcessesByCPU(5)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("%-10s %-8s %s\n", "PID", "CPU%", "COMMAND")
		fmt.Println("-------------------------------")
		for _, p := range procs {
			fmt.Printf("%-10s %-8s %s\n", p.PID, p.CPU, p.Command)
		}
	},
}

var processesMemCmd = &cobra.Command{
	Use:   "memory",
	Short: "Show top 5 processes by memory usage",
	Run: func(cmd *cobra.Command, args []string) {
		procs, err := system.GetTopProcessesByMemory(5)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("%-10s %-8s %s\n", "PID", "MEM%", "COMMAND")
		fmt.Println("-------------------------------")
		for _, p := range procs {
			fmt.Printf("%-10s %-8s %s\n", p.PID, p.Memory, p.Command)
		}
	},
}

func init() {
	processesCmd.AddCommand(processesCPUCmd)
	processesCmd.AddCommand(processesMemCmd)
	rootCmd.AddCommand(processesCmd)
}