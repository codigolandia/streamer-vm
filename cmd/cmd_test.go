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
	if _, err := os.Stat(overlayDiskPath); !os.IsNotExist(err) {
		t.Fatalf("Overlay disk should NOT exist before commit")
	}

	// Test update command: attach ISO
	fakeISO := filepath.Join(tempHome, "test.iso")
	_ = os.WriteFile(fakeISO, []byte("iso-data"), 0644)
	if err := runUpdate([]string{"-iso", fakeISO, "e2e-vm"}); err != nil {
		t.Fatalf("update command failed to attach ISO: %v", err)
	}

	// Test update command: remove ISO
	if err := runUpdate([]string{"-remove-iso", "e2e-vm"}); err != nil {
		t.Fatalf("update command failed to remove ISO: %v", err)
	}

	// Test reset before commit (should fail because no overlay exists)
	if err := runReset([]string{"e2e-vm"}); err == nil {
		t.Fatalf("reset command should have failed before initial commit")
	}

	// Test commit command (initial commit: creates overlay and backups UEFI vars)
	if err := runCommit([]string{"e2e-vm"}); err != nil {
		t.Fatalf("commit command failed: %v", err)
	}

	if _, err := os.Stat(overlayDiskPath); err != nil {
		t.Fatalf("Overlay disk missing after commit: %v", err)
	}

	baseVarsPath := filepath.Join(tempHome, "configs", "e2e-vm", "ovf-vars.base.fd")
	if _, err := os.Stat(baseVarsPath); err != nil {
		t.Fatalf("Base OVMF vars missing after commit: %v", err)
	}

	// Test subsequent commit with -remove-iso
	if err := runCommit([]string{"-remove-iso", "e2e-vm"}); err != nil {
		t.Fatalf("subsequent commit command failed: %v", err)
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

	// Test reset command after commit
	if err := runReset([]string{"e2e-vm"}); err != nil {
		t.Fatalf("reset command failed after commit: %v", err)
	}

	// Test delete command
	if err := runDelete([]string{"-force", "e2e-vm"}); err != nil {
		t.Fatalf("delete command failed: %v", err)
	}
}

func TestPrintHelp(t *testing.T) {
	PrintHelp()
}

func TestExtractLang(t *testing.T) {
	cases := []struct {
		input        []string
		expectedLang string
		expectedArgs []string
	}{
		{
			input:        []string{"--lang", "pt", "status", "myvm"},
			expectedLang: "pt",
			expectedArgs: []string{"status", "myvm"},
		},
		{
			input:        []string{"--lang=pt", "status"},
			expectedLang: "pt",
			expectedArgs: []string{"status"},
		},
		{
			input:        []string{"-lang", "en", "list"},
			expectedLang: "en",
			expectedArgs: []string{"list"},
		},
		{
			input:        []string{"-lang=en", "commit", "myvm"},
			expectedLang: "en",
			expectedArgs: []string{"commit", "myvm"},
		},
		{
			input:        []string{"status", "myvm"},
			expectedLang: "",
			expectedArgs: []string{"status", "myvm"},
		},
	}

	for _, tc := range cases {
		lang, args := extractLang(tc.input)
		if lang != tc.expectedLang {
			t.Errorf("extractLang(%v): expected lang %s, got %s", tc.input, tc.expectedLang, lang)
		}
		if len(args) != len(tc.expectedArgs) {
			t.Errorf("extractLang(%v): expected args %v, got %v", tc.input, tc.expectedArgs, args)
		}
	}
}
