package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Show disk usage per mount point",
	Run: func(cmd *cobra.Command, args []string) {
		disks, err := system.GetDiskMounts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(disks, "", "  ")
			fmt.Println(string(out))
			return
		}
		fmt.Printf("%-20s %-10s %-10s %-10s %s\n", "MOUNT", "TOTAL", "USED", "FREE", "USE%")
		fmt.Println("------------------------------------------------------")
		for _, d := range disks {
			fmt.Printf("%-20s %-10s %-10s %-10s %s\n", d.Mount, d.Total, d.Used, d.Free, system.Colorize(d.Percent, 70, 90))
		}
	},
}

func init() {
	rootCmd.AddCommand(diskCmd)
}