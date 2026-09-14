package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "start",
		ShortKey: "cmd.start.short",
		LongKey:  "cmd.start.long",
		Short:    "Start a virtual machine",
		Long:     "Launches QEMU in background daemon mode and waits for the SPICE display server to be ready.",
		Run:      runStart,
	})
}

func runStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	isoFlag := fs.String("iso", "", "Path to guest OS installation ISO image")
	removeISO := fs.Bool("remove-iso", false, "Remove installation ISO image before starting")
	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	if len(positional) < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm start <name> [-iso <path>] [-remove-iso]")
	}

	if *removeISO && *isoFlag != "" {
		return fmt.Errorf("cannot specify both -iso and -remove-iso")
	}

	name := positional[0]
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if *removeISO {
		cfg.ISOPath = ""
		_ = vm.SaveVMConfig(home, cfg)
	} else if *isoFlag != "" {
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
		internal.Info(i18n.T("msg.already_running", name, pid))
		internal.Info(i18n.T("msg.spice_url", spiceURL))
		return nil
	}

	internal.Info(i18n.T("msg.starting_vm", name))
	if err := vm.StartVM(home, cfg); err != nil {
		return err
	}

	internal.Info(i18n.T("msg.waiting_spice"))
	if vm.WaitUntilSpiceReady(spiceSock, "127.0.0.1", cfg.SpicePort, 15*time.Second) {
		internal.Success(i18n.T("msg.started_success", name))
		internal.Success(i18n.T("msg.spice_url", spiceURL))
		internal.Info(i18n.T("msg.connect_hint", spiceURL))
	} else {
		pid := vm.GetVMPID(home, name)
		internal.Warn(i18n.T("msg.spice_timeout", pid))
		internal.Info(i18n.T("msg.check_log", vm.GetLogFilePath(home, name)))
	}

	return nil
}
