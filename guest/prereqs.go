package guest

import (
	"fmt"
	"os"
	"os/exec"

	"streamer-vm/internal"
)

// PrereqCheck represents the status of a single prerequisite check.
type PrereqCheck struct {
	Name    string
	Passed  bool
	Message string
	Details string
}

// PrereqReport holds all prerequisite check results.
type PrereqReport struct {
	Checks []PrereqCheck
}

// AllPassed returns true if all critical prerequisite checks passed.
func (r *PrereqReport) AllPassed() bool {
	for _, c := range r.Checks {
		if !c.Passed {
			return false
		}
	}
	return true
}

// PrintReport logs each prerequisite check result using the internal logger.
func (r *PrereqReport) PrintReport() {
	for _, c := range r.Checks {
		if c.Passed {
			internal.Success("%s: %s", c.Name, c.Message)
		} else {
			internal.Warn("%s: %s (%s)", c.Name, c.Message, c.Details)
		}
	}
}

// CheckKVM verifies that /dev/kvm exists and is accessible.
func CheckKVM() PrereqCheck {
	info, err := os.Stat("/dev/kvm")
	if err != nil {
		return PrereqCheck{
			Name:    "KVM Acceleration",
			Passed:  false,
			Message: "/dev/kvm not found",
			Details: "Ensure KVM module is loaded (modprobe kvm_amd or kvm_intel)",
		}
	}
	_ = info
	// Check read/write permission
	f, err := os.OpenFile("/dev/kvm", os.O_RDWR, 0)
	if err != nil {
		return PrereqCheck{
			Name:    "KVM Acceleration",
			Passed:  false,
			Message: "/dev/kvm permission denied",
			Details: "Add user to 'kvm' group: sudo usermod -aG kvm $USER",
		}
	}
	f.Close()
	return PrereqCheck{
		Name:    "KVM Acceleration",
		Passed:  true,
		Message: "/dev/kvm accessible",
	}
}

// CheckGPU verifies GPU DRM render node accessibility for VirGL acceleration.
func CheckGPU() PrereqCheck {
	nodes := []string{"/dev/dri/renderD128", "/dev/dri/card0"}
	for _, node := range nodes {
		if _, err := os.Stat(node); err == nil {
			f, err := os.OpenFile(node, os.O_RDWR, 0)
			if err == nil {
				f.Close()
				return PrereqCheck{
					Name:    "GPU DRM (VirGL)",
					Passed:  true,
					Message: fmt.Sprintf("DRM node %s accessible", node),
				}
			}
		}
	}
	return PrereqCheck{
		Name:    "GPU DRM (VirGL)",
		Passed:  false,
		Message: "No writable DRM render node found in /dev/dri/",
		Details: "Ensure user is in 'render' or 'video' group",
	}
}

// CheckQEMU checks if qemu-system-x86_64 is installed and executable.
func CheckQEMU() PrereqCheck {
	path, err := exec.LookPath("qemu-system-x86_64")
	if err != nil {
		return PrereqCheck{
			Name:    "QEMU Binary",
			Passed:  false,
			Message: "qemu-system-x86_64 not found in PATH",
			Details: "Install qemu-system-x86 (e.g. apt install qemu-system-x86)",
		}
	}
	return PrereqCheck{
		Name:    "QEMU Binary",
		Passed:  true,
		Message: fmt.Sprintf("Found %s", path),
	}
}

// CheckQEMUImg checks if qemu-img is installed and executable.
func CheckQEMUImg() PrereqCheck {
	path, err := exec.LookPath("qemu-img")
	if err != nil {
		return PrereqCheck{
			Name:    "QEMU Disk Utility",
			Passed:  false,
			Message: "qemu-img not found in PATH",
			Details: "Install qemu-utils (e.g. apt install qemu-utils)",
		}
	}
	return PrereqCheck{
		Name:    "QEMU Disk Utility",
		Passed:  true,
		Message: fmt.Sprintf("Found %s", path),
	}
}

// StandardOVMFPaths for OVMF_CODE_4M.fd and OVMF_VARS_4M.fd
var StandardOVMFPaths = []struct {
	Code string
	Vars string
}{
	{"/usr/share/OVMF/OVMF_CODE_4M.fd", "/usr/share/OVMF/OVMF_VARS_4M.fd"},
	{"/usr/share/ovmf/OVMF_CODE_4M.fd", "/usr/share/ovmf/OVMF_VARS_4M.fd"},
	{"/usr/share/edk2/x64/OVMF_CODE_4M.fd", "/usr/share/edk2/x64/OVMF_VARS_4M.fd"},
	{"/usr/share/qemu/OVMF_CODE_4M.fd", "/usr/share/qemu/OVMF_VARS_4M.fd"},
	{"/usr/share/OVMF/OVMF_CODE.fd", "/usr/share/OVMF/OVMF_VARS.fd"},
	{"/usr/share/ovmf/OVMF.fd", "/usr/share/ovmf/OVMF_VARS.fd"},
}

// FindOVMFPaths locates UEFI OVMF firmware code and vars templates.
func FindOVMFPaths() (string, string, error) {
	for _, p := range StandardOVMFPaths {
		if _, err1 := os.Stat(p.Code); err1 == nil {
			if _, err2 := os.Stat(p.Vars); err2 == nil {
				return p.Code, p.Vars, nil
			}
		}
	}
	return "", "", fmt.Errorf("OVMF UEFI firmware files not found")
}

// CheckOVMF verifies UEFI OVMF firmware availability.
func CheckOVMF() PrereqCheck {
	code, vars, err := FindOVMFPaths()
	if err != nil {
		return PrereqCheck{
			Name:    "OVMF UEFI Firmware",
			Passed:  false,
			Message: "OVMF_CODE_4M.fd / OVMF_VARS_4M.fd missing",
			Details: "Install ovmf package (e.g. apt install ovmf)",
		}
	}
	return PrereqCheck{
		Name:    "OVMF UEFI Firmware",
		Passed:  true,
		Message: fmt.Sprintf("Found %s and %s", code, vars),
	}
}

// CheckPipewire checks for Pipewire availability on host.
func CheckPipewire() PrereqCheck {
	// Check pipewire runtime socket or binary
	if _, err := exec.LookPath("pipewire"); err == nil {
		return PrereqCheck{
			Name:    "Pipewire Audio",
			Passed:  true,
			Message: "Pipewire binary found",
		}
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		if _, err := os.Stat(fmt.Sprintf("%s/pipewire-0", runtimeDir)); err == nil {
			return PrereqCheck{
				Name:    "Pipewire Audio",
				Passed:  true,
				Message: "Pipewire runtime socket found",
			}
		}
	}
	return PrereqCheck{
		Name:    "Pipewire Audio",
		Passed:  false,
		Message: "Pipewire audio not detected",
		Details: "QEMU pipewire audiodev might fail if PipeWire is not running",
	}
}

// CheckSpiceClient checks if a spice client GUI tool is installed.
func CheckSpiceClient() PrereqCheck {
	clients := []string{"spicy", "virt-viewer", "spice-client-gtk"}
	for _, client := range clients {
		if path, err := exec.LookPath(client); err == nil {
			return PrereqCheck{
				Name:    "Spice Client GUI",
				Passed:  true,
				Message: fmt.Sprintf("Found %s (%s)", client, path),
			}
		}
	}
	return PrereqCheck{
		Name:    "Spice Client GUI",
		Passed:  false,
		Message: "No Spice viewer found (spicy / virt-viewer / spice-client-gtk)",
		Details: "Install spice-client-gtk or virt-viewer (e.g. apt install spice-client-gtk)",
	}
}

// CheckAllPrereqs executes all host prerequisite checks.
func CheckAllPrereqs() PrereqReport {
	return PrereqReport{
		Checks: []PrereqCheck{
			CheckKVM(),
			CheckGPU(),
			CheckQEMU(),
			CheckQEMUImg(),
			CheckOVMF(),
			CheckPipewire(),
			CheckSpiceClient(),
		},
	}
}
