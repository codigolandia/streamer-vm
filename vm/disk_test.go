package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirs(t *testing.T) {
	tempHome := t.TempDir()
	if err := EnsureDirs(tempHome); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	dirs := []string{"configs", "disks", "iso", "state", "logs"}
	for _, dir := range dirs {
		p := filepath.Join(tempHome, dir)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("Directory %s was not created", p)
		}
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "src.txt")
	dst := filepath.Join(tempDir, "dst.txt")

	content := []byte("hello streamer-vm")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("Failed to write src file: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	dstContent, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read dst file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("CopyFile content mismatch: expected %s, got %s", content, dstContent)
	}
}

func TestOverlayDiskLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	baseDisk := filepath.Join(tempDir, "test-base.qcow2")
	overlayDisk := filepath.Join(tempDir, "test-overlay.qcow2")

	// Create base
	if err := CreateBaseDisk(baseDisk, 1); err != nil {
		t.Skipf("qemu-img create failed (possibly not installed): %v", err)
	}

	// Create overlay
	if err := CreateOverlayDisk(baseDisk, overlayDisk); err != nil {
		t.Fatalf("CreateOverlayDisk failed: %v", err)
	}

	// Commit overlay
	if err := CommitOverlayDisk(overlayDisk); err != nil {
		t.Fatalf("CommitOverlayDisk failed: %v", err)
	}

	// Remove overlay
	if err := RemoveOverlayDisk(overlayDisk); err != nil {
		t.Fatalf("RemoveOverlayDisk failed: %v", err)
	}

	if _, err := os.Stat(overlayDisk); !os.IsNotExist(err) {
		t.Errorf("Overlay disk still exists after RemoveOverlayDisk")
	}
}
