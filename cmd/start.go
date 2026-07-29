package cmd

import (
	"flag"
	"fmt"
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
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm start <name>")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if vm.IsVMRunning(home, name) {
		pid := vm.GetVMPID(home, name)
		internal.Info("VM '%s' is already running (PID: %d)", name, pid)
		internal.Info("Spice URL: %s", vm.GetSpiceURL("127.0.0.1", cfg.SpicePort))
		return nil
	}

	internal.Info("Starting VM '%s'...", name)
	if err := vm.StartVM(home, cfg); err != nil {
		return err
	}

	internal.Info("Waiting for Spice server on port %d to become ready...", cfg.SpicePort)
	if vm.WaitUntilSpiceReady("127.0.0.1", cfg.SpicePort, 15*time.Second) {
		spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort)
		internal.Success("VM '%s' started successfully!", name)
		internal.Success("Spice URL: %s", spiceURL)
		internal.Info("Connect using spice-client-gtk (or spicy) or open window capture in OBS Studio.")
	} else {
		pid := vm.GetVMPID(home, name)
		internal.Warn("QEMU started (PID %d), but Spice server port %d did not respond within timeout.", pid, cfg.SpicePort)
		internal.Info("Check log file at: %s", vm.GetLogFilePath(home, name))
	}

	return nil
}
