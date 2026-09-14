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
		Name:     "commit",
		ShortKey: "cmd.commit.short",
		LongKey:  "cmd.commit.long",
		Short:    "Commit VM changes to base image and prepare overlay",
		Long:     "Stops the VM if running, commits current overlay changes (or initializes first overlay if in setup mode) to the base image, saves UEFI vars, and optionally removes the installation ISO.",
		Run:      runCommit,
	})
}

func runCommit(args []string) error {
	fs := flag.NewFlagSet("commit", flag.ContinueOnError)
	removeISO := fs.Bool("remove-iso", false, "Remove installation ISO from VM configuration")
	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	if len(positional) < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm commit <name> [--remove-iso]")
	}

	name := positional[0]
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if vm.IsVMRunning(home, name) {
		internal.Info(i18n.T("msg.stopping_before_commit", name))
		if err := vm.StopVM(home, cfg, 10*time.Second); err != nil {
			return fmt.Errorf("failed to stop VM before commit: %w", err)
		}
	}

	baseDisk := vm.GetBaseDiskPath(home, name)
	if _, err := os.Stat(baseDisk); err != nil {
		return fmt.Errorf("base disk for VM '%s' missing at %s: %w", name, baseDisk, err)
	}

	overlayDisk := vm.GetOverlayDiskPath(home, name)

	if vm.OverlayExists(home, name) {
		internal.Info(i18n.T("msg.committing_overlay", name))
		if err := vm.CommitOverlayDisk(overlayDisk); err != nil {
			return err
		}
		_ = vm.RemoveOverlayDisk(overlayDisk)
		if err := vm.CreateOverlayDisk(baseDisk, overlayDisk); err != nil {
			return fmt.Errorf("failed to recreate clean overlay disk: %w", err)
		}
	} else {
		internal.Info(i18n.T("msg.initial_commit", name))
		if err := vm.CreateOverlayDisk(baseDisk, overlayDisk); err != nil {
			return fmt.Errorf("failed to create overlay disk: %w", err)
		}
	}

	// Backup current OVMF vars as base vars
	ovfVars := vm.GetOVMFVarsPath(home, name)
	if _, err := os.Stat(ovfVars); err == nil {
		baseOVFVars := vm.GetBaseOVMFVarsPath(home, name)
		if err := vm.CopyFile(ovfVars, baseOVFVars); err != nil {
			internal.Warn("Failed to backup UEFI vars: %v", err)
		} else {
			internal.Info(i18n.T("msg.backup_uefi", baseOVFVars))
		}
	}

	if *removeISO {
		cfg.ISOPath = ""
		internal.Info(i18n.T("msg.removed_iso"))
	}

	cfg.State = "stopped"
	if err := vm.SaveVMConfig(home, cfg); err != nil {
		return fmt.Errorf("failed to save VM config: %w", err)
	}

	internal.Success(i18n.T("msg.commit_success", name))
	return nil
}
