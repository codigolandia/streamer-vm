package vm

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

	args, err := BuildQEMUArgs(tempHome, cfg)
	if err != nil {
		t.Fatalf("BuildQEMUArgs failed: %v", err)
	}

	argStr := ""
	for _, a := range args {
		argStr += a + " "
	}

	if !contains(args, "-enable-kvm") {
		t.Errorf("Expected -enable-kvm in args")
	}
	if !contains(args, "spice-app,gl=on") {
		t.Errorf("Expected spice-app,gl=on in args")
	}
	if !contains(args, "port=5900,disable-ticketing=on,addr=127.0.0.1") {
		t.Errorf("Expected spice port config in args")
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
