package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VMConfig stores metadata for a streamer VM.
type VMConfig struct {
	Name      string    `json:"name"`
	CPUs      int       `json:"cpus"`
	MemoryGB  int       `json:"memory_gb"`
	DiskGB    int       `json:"disk_gb"`
	SpicePort int       `json:"spice_port"`
	ISOPath   string    `json:"iso_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	State     string    `json:"state"` // "running", "stopped"
}

// GetStreamerHome returns the root directory for streamer-vm.
// Priority: $STREAMER_HOME > $HOME/.local/share/streamer-vm.
func GetStreamerHome() string {
	if h := os.Getenv("STREAMER_HOME"); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "streamer-vm")
}

// Directory path helpers
func GetConfigsDir(home string) string {
	return filepath.Join(home, "configs")
}

func GetVMConfigDir(home, name string) string {
	return filepath.Join(home, "configs", name)
}

func GetVMConfigPath(home, name string) string {
	return filepath.Join(home, "configs", name, "vm.json")
}

func GetOVMFVarsPath(home, name string) string {
	return filepath.Join(home, "configs", name, "ovf-vars.fd")
}

func GetBaseOVMFVarsPath(home, name string) string {
	return filepath.Join(home, "configs", name, "ovf-vars.base.fd")
}

func GetDisksDir(home string) string {
	return filepath.Join(home, "disks")
}

func GetBaseDiskPath(home, name string) string {
	return filepath.Join(home, "disks", fmt.Sprintf("%s.qcow2", name))
}

func GetOverlayDiskPath(home, name string) string {
	return filepath.Join(home, "disks", fmt.Sprintf("%s-overlay.qcow2", name))
}

func GetISODir(home string) string {
	return filepath.Join(home, "iso")
}

func GetStateDir(home string) string {
	return filepath.Join(home, "state")
}

func GetPIDFilePath(home, name string) string {
	return filepath.Join(home, "state", fmt.Sprintf("%s.pid", name))
}

func GetSpicePortFilePath(home, name string) string {
	return filepath.Join(home, "state", fmt.Sprintf("%s.spice-port", name))
}

func GetSpiceSocketPath(home, name string) string {
	return filepath.Join(home, "state", fmt.Sprintf("%s.spice.sock", name))
}

func GetMonitorSocketPath(home, name string) string {
	return filepath.Join(home, "state", fmt.Sprintf("%s.monitor", name))
}

func GetLogsDir(home string) string {
	return filepath.Join(home, "logs")
}

func GetLogFilePath(home, name string) string {
	return filepath.Join(home, "logs", fmt.Sprintf("%s.log", name))
}

func GetTemplateOVMFVarsPath(home string) string {
	return filepath.Join(home, "template-ovf-vars.fd")
}

// SaveVMConfig writes the VMConfig to configs/<name>/vm.json.
func SaveVMConfig(home string, cfg *VMConfig) error {
	dir := GetVMConfigDir(home, cfg.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create vm config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal vm config: %w", err)
	}
	path := GetVMConfigPath(home, cfg.Name)
	return os.WriteFile(path, data, 0644)
}

// LoadVMConfig reads and parses configs/<name>/vm.json.
func LoadVMConfig(home, name string) (*VMConfig, error) {
	path := GetVMConfigPath(home, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("vm '%s' does not exist or config cannot be read: %w", name, err)
	}
	var cfg VMConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file for vm '%s': %w", name, err)
	}
	return &cfg, nil
}

// VMExists checks if a VM configuration exists.
func VMExists(home, name string) bool {
	path := GetVMConfigPath(home, name)
	_, err := os.Stat(path)
	return err == nil
}

// OverlayExists checks if the overlay disk for a VM exists.
func OverlayExists(home, name string) bool {
	path := GetOverlayDiskPath(home, name)
	_, err := os.Stat(path)
	return err == nil
}

// ListVMs lists all configured VMs.
func ListVMs(home string) ([]*VMConfig, error) {
	configsDir := GetConfigsDir(home)
	entries, err := os.ReadDir(configsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*VMConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read configs directory: %w", err)
	}

	var vms []*VMConfig
	for _, entry := range entries {
		if entry.IsDir() {
			cfg, err := LoadVMConfig(home, entry.Name())
			if err == nil {
				vms = append(vms, cfg)
			}
		}
	}
	return vms, nil
}

// DeleteVM removes a VM configuration and disk files.
func DeleteVM(home, name string, force bool) error {
	if !VMExists(home, name) {
		return fmt.Errorf("vm '%s' does not exist", name)
	}

	// Remove config dir
	configDir := GetVMConfigDir(home, name)
	_ = os.RemoveAll(configDir)

	// Remove base and overlay disks
	_ = os.Remove(GetBaseDiskPath(home, name))
	_ = os.Remove(GetOverlayDiskPath(home, name))

	// Remove state and log files
	_ = os.Remove(GetPIDFilePath(home, name))
	_ = os.Remove(GetSpicePortFilePath(home, name))
	_ = os.Remove(GetMonitorSocketPath(home, name))
	_ = os.Remove(GetLogFilePath(home, name))

	return nil
}
