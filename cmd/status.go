package cmd

import (
	"flag"
	"fmt"

	"streamer-vm/internal"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:  "status",
		Short: "Show status of a specific virtual machine or all VMs",
		Long:  "Displays running status, PID, Spice port, CPUs, RAM, and disk information for a VM.",
		Run:   runStatus,
	})
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	home := vm.GetStreamerHome()

	if fs.NArg() == 1 {
		name := fs.Arg(0)
		cfg, err := vm.LoadVMConfig(home, name)
		if err != nil {
			return err
		}

		running := vm.IsVMRunning(home, name)
		stateStr := "STOPPED"
		pidStr := "N/A"
		if running {
			stateStr = "RUNNING"
			pidStr = fmt.Sprintf("%d", vm.GetVMPID(home, name))
		}

		spiceSock := vm.GetSpiceSocketPath(home, name)
		spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort, spiceSock)

		internal.Info("=== VM Status: %s ===", cfg.Name)
		fmt.Printf("State:      %s\n", stateStr)
		fmt.Printf("PID:        %s\n", pidStr)
		fmt.Printf("CPUs:       %d\n", cfg.CPUs)
		fmt.Printf("Memory:     %d GB\n", cfg.MemoryGB)
		fmt.Printf("Disk:       %d GB\n", cfg.DiskGB)
		fmt.Printf("Spice URL:  %s\n", spiceURL)
		if cfg.ISOPath != "" {
			fmt.Printf("ISO:        %s\n", cfg.ISOPath)
		}
		fmt.Printf("Created:    %s\n", cfg.CreatedAt.Local().Format("2006-01-02 15:04:05"))
		fmt.Printf("Log File:   %s\n", vm.GetLogFilePath(home, name))
		return nil
	}

	vms, err := vm.ListVMs(home)
	if err != nil {
		return err
	}

	if len(vms) == 0 {
		internal.Info("No virtual machines found. Create one with: streamer-vm create <name>")
		return nil
	}

	internal.Info("=== Virtual Machines ===")
	for _, cfg := range vms {
		running := vm.IsVMRunning(home, cfg.Name)
		stateStr := "STOPPED"
		if running {
			stateStr = fmt.Sprintf("RUNNING (PID: %d)", vm.GetVMPID(home, cfg.Name))
		}
		fmt.Printf("- %-15s | State: %-18s | CPUs: %d | RAM: %dGB | Disk: %dGB\n",
			cfg.Name, stateStr, cfg.CPUs, cfg.MemoryGB, cfg.DiskGB)
	}

	return nil
}
