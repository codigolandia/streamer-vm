package guest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
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
	name := i18n.T("prereqs.kvm.name")
	_, err := os.Stat("/dev/kvm")
	if err != nil {
		return PrereqCheck{
			Name:    name,
			Passed:  false,
			Message: i18n.T("prereqs.kvm.not_found"),
			Details: i18n.T("prereqs.kvm.not_found_details"),
		}
	}

	f, err := os.OpenFile("/dev/kvm", os.O_RDWR, 0)
	if err != nil {
		return PrereqCheck{
			Name:    name,
			Passed:  false,
			Message: i18n.T("prereqs.kvm.perm_denied"),
			Details: i18n.T("prereqs.kvm.perm_details"),
		}
	}
	f.Close()
	return PrereqCheck{
		Name:    name,
		Passed:  true,
		Message: i18n.T("prereqs.kvm.accessible"),
	}
}

// CheckGPU verifies GPU DRM render node accessibility for VirGL acceleration.
func CheckGPU() PrereqCheck {
	name := i18n.T("prereqs.gpu.name")
	entries, err := os.ReadDir("/dev/dri")
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "renderD") || strings.HasPrefix(entry.Name(), "card") {
				node := filepath.Join("/dev/dri", entry.Name())
				f, err := os.OpenFile(node, os.O_RDWR, 0)
				if err == nil {
					f.Close()
					return PrereqCheck{
						Name:    name,
						Passed:  true,
						Message: i18n.T("prereqs.gpu.accessible", node),
					}
				}
			}
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  false,
		Message: i18n.T("prereqs.gpu.not_found"),
		Details: i18n.T("prereqs.gpu.details"),
	}
}

// CheckQEMU checks if qemu-system-x86_64 is installed and executable.
func CheckQEMU() PrereqCheck {
	name := i18n.T("prereqs.qemu.name")
	path, err := exec.LookPath("qemu-system-x86_64")
	if err != nil {
		return PrereqCheck{
			Name:    name,
			Passed:  false,
			Message: i18n.T("prereqs.qemu.not_found"),
			Details: i18n.T("prereqs.qemu.details"),
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  true,
		Message: i18n.T("prereqs.qemu.found", path),
	}
}

// CheckQEMUImg checks if qemu-img is installed and executable.
func CheckQEMUImg() PrereqCheck {
	name := i18n.T("prereqs.qemu_img.name")
	path, err := exec.LookPath("qemu-img")
	if err != nil {
		return PrereqCheck{
			Name:    name,
			Passed:  false,
			Message: i18n.T("prereqs.qemu_img.not_found"),
			Details: i18n.T("prereqs.qemu_img.details"),
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  true,
		Message: i18n.T("prereqs.qemu_img.found", path),
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
	name := i18n.T("prereqs.ovmf.name")
	code, vars, err := FindOVMFPaths()
	if err != nil {
		return PrereqCheck{
			Name:    name,
			Passed:  false,
			Message: i18n.T("prereqs.ovmf.missing"),
			Details: i18n.T("prereqs.ovmf.details"),
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  true,
		Message: i18n.T("prereqs.ovmf.found", code, vars),
	}
}

// CheckPipewire checks for Pipewire availability on host.
func CheckPipewire() PrereqCheck {
	name := i18n.T("prereqs.pipewire.name")
	if _, err := exec.LookPath("pipewire"); err == nil {
		return PrereqCheck{
			Name:    name,
			Passed:  true,
			Message: i18n.T("prereqs.pipewire.binary_found"),
		}
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		if _, err := os.Stat(fmt.Sprintf("%s/pipewire-0", runtimeDir)); err == nil {
			return PrereqCheck{
				Name:    name,
				Passed:  true,
				Message: i18n.T("prereqs.pipewire.socket_found"),
			}
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  false,
		Message: i18n.T("prereqs.pipewire.not_found"),
		Details: i18n.T("prereqs.pipewire.details"),
	}
}

// CheckSpiceClient checks if a spice client GUI tool is installed.
func CheckSpiceClient() PrereqCheck {
	name := i18n.T("prereqs.spice.name")
	clients := []string{"spicy", "virt-viewer", "spice-client-gtk"}
	for _, client := range clients {
		if path, err := exec.LookPath(client); err == nil {
			return PrereqCheck{
				Name:    name,
				Passed:  true,
				Message: i18n.T("prereqs.spice.found", client, path),
			}
		}
	}
	return PrereqCheck{
		Name:    name,
		Passed:  false,
		Message: i18n.T("prereqs.spice.not_found"),
		Details: i18n.T("prereqs.spice.details"),
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
