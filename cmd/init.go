package cmd

import (
	"flag"

	"streamer-vm/guest"
	"streamer-vm/internal"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:  "init",
		Short: "Initialize streamer-vm environment and verify host prerequisites",
		Long:  "Creates required directory structure under STREAMER_HOME and verifies host prerequisites (KVM, GPU DRM, QEMU, OVMF firmware).",
		Run:   runInit,
	})
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	home := vm.GetStreamerHome()
	internal.Info("Initializing streamer-vm environment at %s", home)

	if err := vm.EnsureDirs(home); err != nil {
		internal.Error("Failed to create directories: %v", err)
		return err
	}
	internal.Success("Directory structure initialized")

	internal.Info("Checking host prerequisites...")
	report := guest.CheckAllPrereqs()
	report.PrintReport()

	// Copy OVMF vars template if available
	_, varsPath, err := guest.FindOVMFPaths()
	if err == nil {
		templatePath := vm.GetTemplateOVMFVarsPath(home)
		if err := vm.CopyFile(varsPath, templatePath); err == nil {
			internal.Success("Saved UEFI OVMF vars template to %s", templatePath)
		}
	}

	if !report.AllPassed() {
		internal.Warn("Some prerequisite checks failed or returned warnings. Please address issues above if VMs fail to start.")
	} else {
		internal.Success("All prerequisite checks passed! streamer-vm is ready to use.")
	}

	return nil
}
