package cmd

import (
	"flag"
	"fmt"
	"time"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "stop",
		ShortKey: "cmd.stop.short",
		LongKey:  "cmd.stop.long",
		Short:    "Stop a running virtual machine",
		Long:     "Gracefully requests ACPI shutdown via QEMU monitor socket, falling back to SIGTERM/SIGKILL if necessary.",
		Run:      runStop,
	})
}

func runStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	stopTimeout := fs.Int("timeout", 30, "Shutdown timeout in seconds before force killing")

	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	if len(positional) < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm stop <name>")
	}

	name := positional[0]
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if !vm.IsVMRunning(home, name) {
		internal.Info(i18n.T("msg.vm_not_running", name))
		cfg.State = "stopped"
		_ = vm.SaveVMConfig(home, cfg)
		return nil
	}

	internal.Info(i18n.T("msg.stopping_vm", name))
	timeout := time.Duration(*stopTimeout) * time.Second
	if err := vm.StopVM(home, cfg, timeout); err != nil {
		return err
	}

	internal.Success(i18n.T("msg.vm_stopped", name))
	return nil
}
