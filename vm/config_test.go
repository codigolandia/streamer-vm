package vm

import (
	"path/filepath"
	"testing"
	"time"
)

func TestVMConfigCRUD(t *testing.T) {
	tempHome := t.TempDir()

	// Ensure dirs
	if err := EnsureDirs(tempHome); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	cfg := &VMConfig{
		Name:      "test-vm",
		CPUs:      2,
		MemoryGB:  4,
		DiskGB:    20,
		SpicePort: 5901,
		ISOPath:   "/tmp/test.iso",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		State:     "stopped",
	}

	// Save
	if err := SaveVMConfig(tempHome, cfg); err != nil {
		t.Fatalf("SaveVMConfig failed: %v", err)
	}

	// Exists
	if !VMExists(tempHome, "test-vm") {
		t.Errorf("VMExists returned false for saved VM")
	}

	// Load
	loaded, err := LoadVMConfig(tempHome, "test-vm")
	if err != nil {
		t.Fatalf("LoadVMConfig failed: %v", err)
	}

	if loaded.Name != cfg.Name || loaded.CPUs != cfg.CPUs || loaded.MemoryGB != cfg.MemoryGB || loaded.DiskGB != cfg.DiskGB || loaded.SpicePort != cfg.SpicePort {
		t.Errorf("Loaded config %+v does not match saved config %+v", loaded, cfg)
	}

	// List
	vms, err := ListVMs(tempHome)
	if err != nil {
		t.Fatalf("ListVMs failed: %v", err)
	}
	if len(vms) != 1 || vms[0].Name != "test-vm" {
		t.Errorf("ListVMs returned unexpected result: %+v", vms)
	}

	// Delete
	if err := DeleteVM(tempHome, "test-vm", false); err != nil {
		t.Fatalf("DeleteVM failed: %v", err)
	}

	if VMExists(tempHome, "test-vm") {
		t.Errorf("VMExists returned true after DeleteVM")
	}
}

func TestPathHelpers(t *testing.T) {
	home := "/home/test/.local/share/streamer-vm"

	if GetConfigsDir(home) != filepath.Join(home, "configs") {
		t.Errorf("Unexpected ConfigsDir")
	}
	if GetVMConfigPath(home, "myvm") != filepath.Join(home, "configs", "myvm", "vm.json") {
		t.Errorf("Unexpected VMConfigPath")
	}
	if GetBaseDiskPath(home, "myvm") != filepath.Join(home, "disks", "myvm.qcow2") {
		t.Errorf("Unexpected BaseDiskPath")
	}
	if GetOverlayDiskPath(home, "myvm") != filepath.Join(home, "disks", "myvm-overlay.qcow2") {
		t.Errorf("Unexpected OverlayDiskPath")
	}
	if GetPIDFilePath(home, "myvm") != filepath.Join(home, "state", "myvm.pid") {
		t.Errorf("Unexpected PIDFilePath")
	}
}
