package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"streamer-vm/internal"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:  "reset",
		Short: "Reset VM overlay disk back to base image state",
		Long:  "Stops the VM if running, removes the overlay qcow2 disk, and recreates a fresh overlay linked to the base disk image.",
		Run:   runReset,
	})
}

func runReset(args []string) error {
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm reset <name>")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if vm.IsVMRunning(home, name) {
		internal.Info("Stopping VM '%s' before resetting disk...", name)
		if err := vm.StopVM(home, cfg, 10*time.Second); err != nil {
			return fmt.Errorf("failed to stop VM before reset: %w", err)
		}
	}

	baseDisk := vm.GetBaseDiskPath(home, name)
	if _, err := os.Stat(baseDisk); err != nil {
		return fmt.Errorf("base disk for VM '%s' missing at %s: %w", name, baseDisk, err)
	}

	overlayDisk := vm.GetOverlayDiskPath(home, name)
	internal.Info("Recreating overlay disk for VM '%s'...", name)
	_ = vm.RemoveOverlayDisk(overlayDisk)

	if err := vm.CreateOverlayDisk(baseDisk, overlayDisk); err != nil {
		return fmt.Errorf("failed to recreate overlay disk: %w", err)
	}

	cfg.State = "stopped"
	_ = vm.SaveVMConfig(home, cfg)

	internal.Success("VM '%s' overlay disk has been reset to clean base state.", name)
	return nil
}
