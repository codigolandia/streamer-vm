package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "reset",
		ShortKey: "cmd.reset.short",
		LongKey:  "cmd.reset.long",
		Short:    "Reset VM overlay disk back to base image state",
		Long:     "Stops the VM if running, removes the overlay qcow2 disk, and recreates a fresh overlay linked to the base disk image.",
		Run:      runReset,
	})
}

func runReset(args []string) error {
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	if len(positional) < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm reset <name>")
	}

	name := positional[0]
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if vm.IsVMRunning(home, name) {
		internal.Info(i18n.T("msg.stopping_before_reset", name))
		if err := vm.StopVM(home, cfg, 10*time.Second); err != nil {
			return fmt.Errorf("failed to stop VM before reset: %w", err)
		}
	}

	baseDisk := vm.GetBaseDiskPath(home, name)
	if _, err := os.Stat(baseDisk); err != nil {
		return fmt.Errorf("base disk for VM '%s' missing at %s: %w", name, baseDisk, err)
	}

	if !vm.OverlayExists(home, name) {
		return fmt.Errorf("VM '%s' has not been committed yet (no overlay disk found); run 'streamer-vm commit %s' after initial setup", name, name)
	}

	overlayDisk := vm.GetOverlayDiskPath(home, name)
	internal.Info(i18n.T("msg.recreating_overlay", name))
	_ = vm.RemoveOverlayDisk(overlayDisk)

	if err := vm.CreateOverlayDisk(baseDisk, overlayDisk); err != nil {
		return fmt.Errorf("failed to recreate overlay disk: %w", err)
	}

	// Restore base OVMF vars if available
	baseOVFVars := vm.GetBaseOVMFVarsPath(home, name)
	if _, err := os.Stat(baseOVFVars); err == nil {
		currentOVFVars := vm.GetOVMFVarsPath(home, name)
		if err := vm.CopyFile(baseOVFVars, currentOVFVars); err != nil {
			internal.Warn("Failed to restore base UEFI vars: %v", err)
		} else {
			internal.Info(i18n.T("msg.restored_uefi", baseOVFVars))
		}
	}

	cfg.State = "stopped"
	_ = vm.SaveVMConfig(home, cfg)

	internal.Success(i18n.T("msg.reset_success", name))
	return nil
}
