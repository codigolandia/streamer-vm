package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"streamer-vm/guest"
	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "create",
		ShortKey: "cmd.create.short",
		LongKey:  "cmd.create.long",
		Short:    "Create a new virtual machine configuration and disk images",
		Long:     "Creates base qcow2 disk, prepares UEFI OVMF variables, and generates vm.json configuration.",
		Run:      runCreate,
	})
}

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	cpus := fs.Int("cpus", 4, "Number of vCPUs")
	memory := fs.Int("memory", 8, "RAM memory in GB")
	disk := fs.Int("disk", 50, "Disk size in GB")
	iso := fs.String("iso", "", "Path to guest OS installation ISO image")
	spicePort := fs.Int("spice-port", 0, "Spice port (default: auto 5900-5999)")
	nameFlag := fs.String("name", "", "VM name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	name := *nameFlag
	if fs.NArg() > 0 && fs.Arg(0) != "" {
		name = fs.Arg(0)
	}
	if name == "" {
		return fmt.Errorf("VM name is required (pass as argument or via -name)")
	}

	home := vm.GetStreamerHome()
	if err := vm.EnsureDirs(home); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	if vm.VMExists(home, name) {
		return fmt.Errorf("VM '%s' already exists", name)
	}

	// Resolve Spice port
	port := *spicePort
	if port <= 0 {
		allocatedPort, err := vm.AllocateFreeSpicePort(5900, 5999)
		if err != nil {
			return fmt.Errorf("failed to allocate free Spice port: %w", err)
		}
		port = allocatedPort
	}

	// Resolve ISO path if provided
	isoPath := *iso
	if isoPath != "" {
		absISO, err := filepath.Abs(isoPath)
		if err == nil {
			isoPath = absISO
		}
		if _, err := os.Stat(isoPath); err != nil {
			internal.Warn("ISO file %s does not exist or is not readable", isoPath)
		}
	}

	internal.Info(i18n.T("msg.creating_vm", name, *cpus, *memory, *disk, port))

	// Create base disk (overlay will be created on first commit)
	baseDisk := vm.GetBaseDiskPath(home, name)
	internal.Info(i18n.T("msg.creating_base_disk", baseDisk, *disk))
	if err := vm.CreateBaseDisk(baseDisk, *disk); err != nil {
		return err
	}

	// Copy OVMF vars template
	vmConfigDir := vm.GetVMConfigDir(home, name)
	if err := os.MkdirAll(vmConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	ovfVarsDst := vm.GetOVMFVarsPath(home, name)
	templateVars := vm.GetTemplateOVMFVarsPath(home)
	varsSrc := templateVars
	if _, err := os.Stat(varsSrc); err != nil {
		_, sysVars, err := guest.FindOVMFPaths()
		if err != nil {
			_ = os.Remove(baseDisk)
			return fmt.Errorf("OVMF vars template not found: %w", err)
		}
		varsSrc = sysVars
	}

	if err := vm.CopyFile(varsSrc, ovfVarsDst); err != nil {
		_ = os.Remove(baseDisk)
		return fmt.Errorf("failed to copy OVMF vars template: %w", err)
	}

	// Save state/port
	portFile := vm.GetSpicePortFilePath(home, name)
	_ = os.WriteFile(portFile, []byte(fmt.Sprintf("%d\n", port)), 0644)

	// Save config
	cfg := &vm.VMConfig{
		Name:      name,
		CPUs:      *cpus,
		MemoryGB:  *memory,
		DiskGB:    *disk,
		SpicePort: port,
		ISOPath:   isoPath,
		CreatedAt: time.Now().UTC(),
		State:     "stopped",
	}

	if err := vm.SaveVMConfig(home, cfg); err != nil {
		return err
	}

	internal.Success(i18n.T("msg.vm_created", name))
	internal.Info(i18n.T("msg.start_hint", name))
	if isoPath != "" {
		internal.Info(i18n.T("msg.commit_hint", name))
	}

	return nil
}
