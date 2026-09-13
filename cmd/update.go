package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "update",
		ShortKey: "cmd.update.short",
		LongKey:  "cmd.update.long",
		Short:    "Update VM configuration (e.g. attach or remove ISO)",
		Long:     "Updates configuration options of an existing virtual machine, such as attaching or removing an ISO image.",
		Run:      runUpdate,
	})
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	isoFlag := fs.String("iso", "", "Path to installation ISO image to attach")
	removeISO := fs.Bool("remove-iso", false, "Detach/remove installation ISO image from VM")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm update <name> [--iso <path> | --remove-iso]")
	}

	if *isoFlag != "" && *removeISO {
		return fmt.Errorf("cannot specify both -iso and -remove-iso")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	changed := false

	if *removeISO {
		if cfg.ISOPath != "" {
			cfg.ISOPath = ""
			changed = true
			internal.Info(i18n.T("msg.iso_removed", name))
		} else {
			internal.Info(i18n.T("msg.no_iso", name))
		}
	} else if *isoFlag != "" {
		absISO, err := filepath.Abs(*isoFlag)
		if err == nil {
			*isoFlag = absISO
		}
		if _, err := os.Stat(*isoFlag); err != nil {
			internal.Warn("ISO file %s does not exist or is not readable", *isoFlag)
		}
		cfg.ISOPath = *isoFlag
		changed = true
		internal.Info(i18n.T("msg.iso_updated", name, *isoFlag))
	}

	if !changed && !*removeISO && *isoFlag == "" {
		return fmt.Errorf("no update options specified; run 'streamer-vm help update' for usage")
	}

	if err := vm.SaveVMConfig(home, cfg); err != nil {
		return fmt.Errorf("failed to save VM config: %w", err)
	}

	internal.Success(i18n.T("msg.vm_updated", name))
	return nil
}
