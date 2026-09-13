package vm

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"streamer-vm/guest"
)

func TestBuildQEMUArgs(t *testing.T) {
	tempHome := t.TempDir()
	_ = EnsureDirs(tempHome)

	cfg := &VMConfig{
		Name:      "test-vm-qemu",
		CPUs:      4,
		MemoryGB:  8,
		DiskGB:    50,
		SpicePort: 5900,
		CreatedAt: time.Now().UTC(),
		State:     "stopped",
	}

	// Create dummy files required for BuildQEMUArgs
	ovfVars := GetOVMFVarsPath(tempHome, cfg.Name)
	_ = os.MkdirAll(filepath.Dir(ovfVars), 0755)
	_ = os.WriteFile(ovfVars, []byte("dummy-vars"), 0644)

	overlayDisk := GetOverlayDiskPath(tempHome, cfg.Name)
	_ = os.WriteFile(overlayDisk, []byte("dummy-disk"), 0644)

	// If OVMF paths are not installed on host, register dummy paths for test
	_, _, err := guest.FindOVMFPaths()
	if err != nil {
		dummyCode := filepath.Join(tempHome, "OVMF_CODE_4M.fd")
		dummyVars := filepath.Join(tempHome, "OVMF_VARS_4M.fd")
		_ = os.WriteFile(dummyCode, []byte("dummy-code"), 0644)
		_ = os.WriteFile(dummyVars, []byte("dummy-vars"), 0644)
		guest.StandardOVMFPaths = append([]struct{ Code, Vars string }{{Code: dummyCode, Vars: dummyVars}}, guest.StandardOVMFPaths...)
	}

	args, err := BuildQEMUArgs(tempHome, cfg)
	if err != nil {
		t.Fatalf("BuildQEMUArgs failed: %v", err)
	}

	if !contains(args, "-enable-kvm") {
		t.Errorf("Expected -enable-kvm in args")
	}
	if !contains(args, "-vga") {
		t.Errorf("Expected -vga in args")
	}
	if !contains(args, "virtio-gpu-gl-pci") {
		t.Errorf("Expected virtio-gpu-gl-pci in args")
	}

	// Test fallback to base disk when overlay is missing
	_ = os.Remove(overlayDisk)
	baseDisk := GetBaseDiskPath(tempHome, cfg.Name)
	_ = os.WriteFile(baseDisk, []byte("dummy-base"), 0644)

	argsBase, err := BuildQEMUArgs(tempHome, cfg)
	if err != nil {
		t.Fatalf("BuildQEMUArgs with base disk failed: %v", err)
	}
	expectedDrive := "id=disk,if=none,file=" + baseDisk + ",format=qcow2,cache=none,aio=native"
	if !contains(argsBase, expectedDrive) {
		t.Errorf("Expected base disk drive in args: %s", expectedDrive)
	}

	// Test error when neither disk exists
	_ = os.Remove(baseDisk)
	_, err = BuildQEMUArgs(tempHome, cfg)
	if err == nil {
		t.Fatalf("Expected error when neither disk exists, got nil")
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
