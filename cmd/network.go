package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zemation/sysinfo/system"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Show network information",
}

var networkInterfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Show network interfaces and traffic stats",
	Run: func(cmd *cobra.Command, args []string) {
		ifaces, err := system.GetNetworkInterfaces()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(ifaces, "", "  ")
			fmt.Println(string(out))
			return
		}
		fmt.Printf("%-12s %-18s %-12s %s\n", "INTERFACE", "IP", "RX", "TX")
		fmt.Println("------------------------------------------------------")
		for _, i := range ifaces {
			fmt.Printf("%-12s %-18s %-12s %s\n", i.Name, i.IP, i.RX, i.TX)
		}
	},
}

var networkPortsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Show listening ports and owning processes",
	Run: func(cmd *cobra.Command, args []string) {
		ports, err := system.GetListeningPorts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(ports, "", "  ")
			fmt.Println(string(out))
			return
		}
		fmt.Printf("%-8s %-8s %-12s %-8s %s\n", "PORT", "PROTO", "STATE", "PID", "COMMAND")
		fmt.Println("------------------------------------------------------")
		for _, p := range ports {
			fmt.Printf("%-8s %-8s %-12s %-8s %s\n", p.Port, p.Proto, p.State, p.PID, p.Command)
		}
	},
}

func init() {
	networkCmd.AddCommand(networkInterfacesCmd)
	networkCmd.AddCommand(networkPortsCmd)
	rootCmd.AddCommand(networkCmd)
}