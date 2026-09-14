package cmd

import (
	"flag"
	"fmt"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "status",
		ShortKey: "cmd.status.short",
		LongKey:  "cmd.status.long",
		Short:    "Show status of a specific virtual machine or all VMs",
		Long:     "Displays running status, PID, Spice port, CPUs, RAM, and disk information for a VM.",
		Run:      runStatus,
	})
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	home := vm.GetStreamerHome()

	if len(positional) == 1 {
		name := positional[0]
		cfg, err := vm.LoadVMConfig(home, name)
		if err != nil {
			return err
		}

		running := vm.IsVMRunning(home, name)
		stateStr := i18n.T("state.stopped")
		pidStr := "N/A"
		if running {
			stateStr = i18n.T("state.running")
			pidStr = fmt.Sprintf("%d", vm.GetVMPID(home, name))
		}

		spiceSock := vm.GetSpiceSocketPath(home, name)
		spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort, spiceSock)

		diskModeStr := i18n.T("status.disk_mode_base")
		if vm.OverlayExists(home, name) {
			diskModeStr = i18n.T("status.disk_mode_overlay")
		}

		isoStr := cfg.ISOPath
		if isoStr == "" {
			isoStr = i18n.T("status.none")
		}

		internal.Info(i18n.T("status.header", cfg.Name))
		fmt.Printf("%-16s %s\n", i18n.T("status.state"), stateStr)
		fmt.Printf("%-16s %s\n", i18n.T("status.pid"), pidStr)
		fmt.Printf("%-16s %d\n", i18n.T("status.cpus"), cfg.CPUs)
		fmt.Printf("%-16s %d GB\n", i18n.T("status.memory"), cfg.MemoryGB)
		fmt.Printf("%-16s %d GB\n", i18n.T("status.disk"), cfg.DiskGB)
		fmt.Printf("%-16s %s\n", i18n.T("status.disk_mode"), diskModeStr)
		fmt.Printf("%-16s %s\n", i18n.T("status.spice_url"), spiceURL)
		fmt.Printf("%-16s %s\n", i18n.T("status.iso"), isoStr)
		fmt.Printf("%-16s %s\n", i18n.T("status.created"), cfg.CreatedAt.Local().Format("2006-01-02 15:04:05"))
		fmt.Printf("%-16s %s\n", i18n.T("status.log_file"), vm.GetLogFilePath(home, name))
		return nil
	}

	vms, err := vm.ListVMs(home)
	if err != nil {
		return err
	}

	if len(vms) == 0 {
		internal.Info(i18n.T("status.no_vms"))
		return nil
	}

	internal.Info(i18n.T("status.header_all"))
	for _, cfg := range vms {
		running := vm.IsVMRunning(home, cfg.Name)
		stateStr := i18n.T("state.stopped")
		if running {
			stateStr = fmt.Sprintf("%s (PID: %d)", i18n.T("state.running"), vm.GetVMPID(home, cfg.Name))
		}
		fmt.Printf("- %-15s | %s %-18s | CPUs: %d | RAM: %dGB | Disk: %dGB\n",
			cfg.Name, i18n.T("status.state"), stateStr, cfg.CPUs, cfg.MemoryGB, cfg.DiskGB)
	}

	return nil
}
