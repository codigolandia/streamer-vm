package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"streamer-vm/internal"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:  "start",
		Short: "Start a virtual machine",
		Long:  "Launches QEMU in background daemon mode and waits for the SPICE display server to be ready.",
		Run:   runStart,
	})
}

func runStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	isoFlag := fs.String("iso", "", "Path to guest OS installation ISO image")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm start <name> [-iso <path>]")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if *isoFlag != "" {
		absISO, err := filepath.Abs(*isoFlag)
		if err == nil {
			isoPath := absISO
			if _, err := os.Stat(isoPath); err != nil {
				internal.Warn("ISO file %s does not exist or is not readable", isoPath)
			}
			cfg.ISOPath = isoPath
			_ = vm.SaveVMConfig(home, cfg)
		}
	}

	spiceSock := vm.GetSpiceSocketPath(home, name)
	spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort, spiceSock)

	if vm.IsVMRunning(home, name) {
		pid := vm.GetVMPID(home, name)
		internal.Info("VM '%s' is already running (PID: %d)", name, pid)
		internal.Info("Spice URL: %s", spiceURL)
		return nil
	}

	internal.Info("Starting VM '%s'...", name)
	if err := vm.StartVM(home, cfg); err != nil {
		return err
	}

	internal.Info("Waiting for Spice server to become ready...")
	if vm.WaitUntilSpiceReady(spiceSock, "127.0.0.1", cfg.SpicePort, 15*time.Second) {
		internal.Success("VM '%s' started successfully!", name)
		internal.Success("Spice URL: %s", spiceURL)
		internal.Info("Connect using: spicy --uri=%s (or virt-viewer)", spiceURL)
	} else {
		pid := vm.GetVMPID(home, name)
		internal.Warn("QEMU started (PID %d), but Spice server did not respond within timeout.", pid)
		internal.Info("Check log file at: %s", vm.GetLogFilePath(home, name))
	}

	return nil
}
