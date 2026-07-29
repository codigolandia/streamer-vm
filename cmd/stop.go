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
		Name:  "stop",
		Short: "Stop a running virtual machine",
		Long:  "Sends an ACPI shutdown signal to QEMU via monitor socket and waits for a graceful shutdown.",
		Run:   runStop,
	})
}

func runStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	stopTimeout := fs.Int("timeout", 30, "Shutdown timeout in seconds before force killing")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm stop <name>")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if !vm.IsVMRunning(home, name) {
		internal.Info("VM '%s' is not running", name)
		cfg.State = "stopped"
		_ = vm.SaveVMConfig(home, cfg)
		return nil
	}

	internal.Info("Stopping VM '%s' (timeout: %ds)...", name, *stopTimeout)
	timeout := time.Duration(*stopTimeout) * time.Second
	if err := vm.StopVM(home, cfg, timeout); err != nil {
		return err
	}

	internal.Success("VM '%s' stopped successfully.", name)
	return nil
}
