package cmd

import (
	"flag"

	"streamer-vm/guest"
	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "init",
		ShortKey: "cmd.init.short",
		LongKey:  "cmd.init.long",
		Short:    "Initialize streamer-vm environment and verify host prerequisites",
		Long:     "Creates required directory structure under STREAMER_HOME and verifies host prerequisites (KVM, GPU DRM, QEMU, OVMF firmware).",
		Run:      runInit,
	})
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	if _, err := ParseAll(fs, args); err != nil {
		return err
	}

	home := vm.GetStreamerHome()
	internal.Info(i18n.T("prereqs.init_env", home))

	if err := vm.EnsureDirs(home); err != nil {
		internal.Error("Failed to create directories: %v", err)
		return err
	}
	internal.Success(i18n.T("prereqs.dir_ok"))

	internal.Info(i18n.T("prereqs.checking"))
	report := guest.CheckAllPrereqs()
	report.PrintReport()

	// Copy OVMF vars template if available
	_, varsPath, err := guest.FindOVMFPaths()
	if err == nil {
		templatePath := vm.GetTemplateOVMFVarsPath(home)
		if err := vm.CopyFile(varsPath, templatePath); err == nil {
			internal.Success(i18n.T("prereqs.saved_vars", templatePath))
		}
	}

	if !report.AllPassed() {
		internal.Warn(i18n.T("prereqs.some_failed"))
	} else {
		internal.Success(i18n.T("prereqs.all_passed"))
	}

	return nil
}
