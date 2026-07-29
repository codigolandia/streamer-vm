package vm

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"streamer-vm/guest"
	"streamer-vm/internal"
)

// BuildQEMUArgs constructs the CLI arguments for starting QEMU.
func BuildQEMUArgs(home string, cfg *VMConfig) ([]string, error) {
	ovmfCode, _, err := guest.FindOVMFPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to locate OVMF firmware: %w", err)
	}

	ovfVars := GetOVMFVarsPath(home, cfg.Name)
	if _, err := os.Stat(ovfVars); err != nil {
		return nil, fmt.Errorf("OVMF vars file missing for VM '%s' at %s", cfg.Name, ovfVars)
	}

	overlayDisk := GetOverlayDiskPath(home, cfg.Name)
	if _, err := os.Stat(overlayDisk); err != nil {
		return nil, fmt.Errorf("overlay disk missing for VM '%s' at %s", cfg.Name, overlayDisk)
	}

	logFile := GetLogFilePath(home, cfg.Name)
	pidFile := GetPIDFilePath(home, cfg.Name)
	monitorSock := GetMonitorSocketPath(home, cfg.Name)

	args := []string{
		"-enable-kvm",
		"-machine", "q35,accel=kvm",
		"-cpu", "host,kvm=on,vendor=GenuineIntel",
		"-smp", fmt.Sprintf("%d,sockets=1,cores=%d,threads=1", cfg.CPUs, cfg.CPUs),
		"-m", fmt.Sprintf("%dG", cfg.MemoryGB),

		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", ovmfCode),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", ovfVars),

		"-vga", "none",
		"-device", "virtio-gpu-gl-pci",
		"-display", "egl-headless",
		"-spice", fmt.Sprintf("port=%d,disable-ticketing=on,addr=127.0.0.1", cfg.SpicePort),

		"-device", "virtio-scsi-pci",
		"-device", "scsi-hd,drive=disk,bootindex=1",
		"-drive", fmt.Sprintf("id=disk,if=none,file=%s,format=qcow2,cache=none,aio=native", overlayDisk),

		"-device", "virtio-net-pci,netdev=net",
		"-netdev", "user,id=net",

		"-audiodev", "pipewire,id=audio",
		"-device", "intel-hda",
		"-device", "hda-output,audiodev=audio",

		"-device", "virtio-serial-pci",
		"-chardev", "spicevmc,id=vdagent,name=vdagent",
		"-device", "virtserialport,chardev=vdagent,name=com.redhat.spice.0",

		"-monitor", fmt.Sprintf("unix:%s,server,nowait", monitorSock),
		"-pidfile", pidFile,

		"-boot", "menu=on",
		"-daemonize",
		"-D", logFile,
	}

	if cfg.ISOPath != "" {
		if _, err := os.Stat(cfg.ISOPath); err == nil {
			args = append(args, "-device", "scsi-cd,drive=cdrom,bootindex=0", "-drive", fmt.Sprintf("id=cdrom,if=none,file=%s,media=cdrom,readonly=on", cfg.ISOPath))
		} else {
			internal.Warn("ISO file %s specified but not found, booting without ISO", cfg.ISOPath)
		}
	}

	return args, nil
}

// GetVMPID reads the PID of the VM from state/<name>.pid. Returns 0 if not found or invalid.
func GetVMPID(home, name string) int {
	pidFile := GetPIDFilePath(home, name)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0
	}
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}
	return pid
}

// IsVMRunning checks whether the QEMU process for the VM is alive.
func IsVMRunning(home, name string) bool {
	pid := GetVMPID(home, name)
	if pid <= 0 {
		return false
	}
	// Check signal 0
	err := syscall.Kill(pid, 0)
	return err == nil
}

// StartVM launches the QEMU process in daemonized mode.
func StartVM(home string, cfg *VMConfig) error {
	if IsVMRunning(home, cfg.Name) {
		return fmt.Errorf("VM '%s' is already running (PID: %d)", cfg.Name, GetVMPID(home, cfg.Name))
	}

	// Remove stale monitor socket or pidfile if present
	_ = os.Remove(GetMonitorSocketPath(home, cfg.Name))
	_ = os.Remove(GetPIDFilePath(home, cfg.Name))

	args, err := BuildQEMUArgs(home, cfg)
	if err != nil {
		return fmt.Errorf("failed to build QEMU command: %w", err)
	}

	cmd := exec.Command("qemu-system-x86_64", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if outStr == "" {
			logData, logErr := os.ReadFile(GetLogFilePath(home, cfg.Name))
			if logErr == nil && len(logData) > 0 {
				outStr = strings.TrimSpace(string(logData))
			}
		}
		return fmt.Errorf("QEMU launch failed: %w (output: %s)", err, outStr)
	}

	// Update state in config
	cfg.State = "running"
	_ = SaveVMConfig(home, cfg)

	return nil
}

// StopVM attempts a graceful ACPI shutdown via QEMU monitor, falling back to SIGTERM/SIGKILL.
func StopVM(home string, cfg *VMConfig, timeout time.Duration) error {
	if !IsVMRunning(home, cfg.Name) {
		cfg.State = "stopped"
		_ = SaveVMConfig(home, cfg)
		_ = os.Remove(GetPIDFilePath(home, cfg.Name))
		_ = os.Remove(GetMonitorSocketPath(home, cfg.Name))
		return nil
	}

	pid := GetVMPID(home, cfg.Name)
	monitorSock := GetMonitorSocketPath(home, cfg.Name)

	// Attempt ACPI shutdown via monitor socket
	sentACPI := false
	conn, err := net.DialTimeout("unix", monitorSock, 2*time.Second)
	if err == nil {
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		_, err = conn.Write([]byte("system_powerdown\n"))
		if err == nil {
			sentACPI = true
			internal.Info("Sent ACPI powerdown command to VM '%s'", cfg.Name)
		}
		_ = conn.Close()
	}

	if !sentACPI {
		internal.Warn("Monitor socket unavailable; sending SIGTERM to VM '%s' (PID %d)", cfg.Name, pid)
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}

	// Wait for process to exit up to timeout
	deadline := time.Now().Add(timeout)
	stopped := false
	for time.Now().Before(deadline) {
		if !IsVMRunning(home, cfg.Name) {
			stopped = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !stopped {
		internal.Warn("VM '%s' did not stop within timeout; force killing (SIGKILL)", cfg.Name)
		_ = syscall.Kill(pid, syscall.SIGKILL)
		time.Sleep(500 * time.Millisecond)
	}

	// Cleanup state
	_ = os.Remove(GetPIDFilePath(home, cfg.Name))
	_ = os.Remove(GetMonitorSocketPath(home, cfg.Name))

	cfg.State = "stopped"
	_ = SaveVMConfig(home, cfg)

	return nil
}
