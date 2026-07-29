package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLICommandsE2E(t *testing.T) {
	if _, err := exec.LookPath("qemu-img"); err != nil {
		t.Skip("qemu-img binary not found in PATH; skipping E2E CLI test")
	}

	tempHome := t.TempDir()
	t.Setenv("STREAMER_HOME", tempHome)

	// Test init
	if err := runInit([]string{}); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Test create command
	createArgs := []string{"-cpus", "2", "-memory", "4", "-disk", "10", "-spice-port", "5910", "e2e-vm"}
	if err := runCreate(createArgs); err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	// Verify files created
	vmConfigPath := filepath.Join(tempHome, "configs", "e2e-vm", "vm.json")
	if _, err := os.Stat(vmConfigPath); err != nil {
		t.Fatalf("VM config file missing at %s: %v", vmConfigPath, err)
	}

	baseDiskPath := filepath.Join(tempHome, "disks", "e2e-vm.qcow2")
	if _, err := os.Stat(baseDiskPath); err != nil {
		t.Fatalf("Base disk missing at %s: %v", baseDiskPath, err)
	}

	overlayDiskPath := filepath.Join(tempHome, "disks", "e2e-vm-overlay.qcow2")
	if _, err := os.Stat(overlayDiskPath); err != nil {
		t.Fatalf("Overlay disk missing at %s: %v", overlayDiskPath, err)
	}

	// Test status command
	if err := runStatus([]string{"e2e-vm"}); err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	// Test status all VMs command
	if err := runStatus([]string{}); err != nil {
		t.Fatalf("status command (all) failed: %v", err)
	}

	// Test list command
	if err := runList([]string{}); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	// Test spice-url command
	if err := runSpiceURL([]string{"e2e-vm"}); err != nil {
		t.Fatalf("spice-url command failed: %v", err)
	}

	// Test reset command
	if err := runReset([]string{"e2e-vm"}); err != nil {
		t.Fatalf("reset command failed: %v", err)
	}

	// Test delete command
	if err := runDelete([]string{"-force", "e2e-vm"}); err != nil {
		t.Fatalf("delete command failed: %v", err)
	}
}

func TestPrintHelp(t *testing.T) {
	PrintHelp()
}
