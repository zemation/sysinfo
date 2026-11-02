package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "Show GPU details, driver version, memory, and utilization",
	Run: func(cmd *cobra.Command, args []string) {
		gpu := system.GetGPUInfo()
		if jsonOutput {
			out, _ := json.MarshalIndent(gpu, "", "  ")
			fmt.Println(string(out))
			return
		}
		fmt.Println("Name:          ", gpu.Name)
		fmt.Println("Driver:        ", gpu.DriverVersion)
		fmt.Println("Memory Total:  ", gpu.MemoryTotal)
		fmt.Println("Memory Used:   ", gpu.MemoryUsed)
		fmt.Println("Utilization:   ", system.Colorize(gpu.Utilization, 70, 90))
	},
}

func init() {
	rootCmd.AddCommand(gpuCmd)
}
