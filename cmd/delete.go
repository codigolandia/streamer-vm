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
		Name:  "delete",
		Short: "Delete a virtual machine configuration and disk images",
		Long:  "Stops the VM if running and permanently deletes its configuration, base disk, overlay disk, state, and logs.",
		Run:   runDelete,
	})
}

func runDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	force := fs.Bool("force", false, "Force deletion of running VM")
	fs.BoolVar(force, "f", false, "Force deletion of running VM (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm delete <name>")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	if !vm.VMExists(home, name) {
		return fmt.Errorf("VM '%s' does not exist", name)
	}

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	if vm.IsVMRunning(home, name) {
		if !*force {
			return fmt.Errorf("VM '%s' is currently running; stop it first using 'streamer-vm stop %s' or pass -force", name, name)
		}
		internal.Info("Force stopping running VM '%s' before deletion...", name)
		_ = vm.StopVM(home, cfg, 5*time.Second)
	}

	internal.Info("Deleting VM '%s'...", name)
	if err := vm.DeleteVM(home, name, *force); err != nil {
		return err
	}

	internal.Success("VM '%s' and associated disk files deleted successfully.", name)
	return nil
}
