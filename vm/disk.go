package vm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// EnsureDirs creates all required directories under STREAMER_HOME.
func EnsureDirs(home string) error {
	dirs := []string{
		GetConfigsDir(home),
		GetDisksDir(home),
		GetISODir(home),
		GetStateDir(home),
		GetLogsDir(home),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// CreateBaseDisk creates a new base qcow2 image with specified size in GB.
func CreateBaseDisk(path string, sizeGB int) error {
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", path, fmt.Sprintf("%dG", sizeGB))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img create base failed: %w (output: %s)", err, string(output))
	}
	return nil
}

// CreateOverlayDisk creates a writable qcow2 overlay image linked to the base image.
func CreateOverlayDisk(basePath, overlayPath string) error {
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-b", basePath, "-F", "qcow2", overlayPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img create overlay failed: %w (output: %s)", err, string(output))
	}
	return nil
}

// RemoveOverlayDisk deletes the overlay disk file.
func RemoveOverlayDisk(overlayPath string) error {
	if err := os.Remove(overlayPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove overlay disk %s: %w", overlayPath, err)
	}
	return nil
}

// CommitOverlayDisk commits changes from the overlay image back into its base backing file.
func CommitOverlayDisk(overlayPath string) error {
	cmd := exec.Command("qemu-img", "commit", overlayPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img commit failed: %w (output: %s)", err, string(output))
	}
	return nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", src, dst, err)
	}

	return out.Sync()
}
