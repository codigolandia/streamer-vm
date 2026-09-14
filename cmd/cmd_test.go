package cmd

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"streamer-vm/vm"
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

	// Test create command with positional name first and flags after (including -iso)
	fakeISO := filepath.Join(tempHome, "test.iso")
	_ = os.WriteFile(fakeISO, []byte("iso-data"), 0644)
	createArgs := []string{"e2e-vm", "-cpus", "2", "-memory", "4", "-disk", "10", "-spice-port", "5910", "-iso", fakeISO}
	if err := runCreate(createArgs); err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	// Verify files created and ISO path stored correctly
	vmConfigPath := filepath.Join(tempHome, "configs", "e2e-vm", "vm.json")
	if _, err := os.Stat(vmConfigPath); err != nil {
		t.Fatalf("VM config file missing at %s: %v", vmConfigPath, err)
	}

	cfg, err := vm.LoadVMConfig(tempHome, "e2e-vm")
	if err != nil {
		t.Fatalf("failed to load VM config: %v", err)
	}
	if cfg.ISOPath != fakeISO {
		t.Fatalf("expected ISO path %q, got %q", fakeISO, cfg.ISOPath)
	}
	if cfg.CPUs != 2 || cfg.MemoryGB != 4 || cfg.DiskGB != 10 || cfg.SpicePort != 5910 {
		t.Fatalf("expected config flags to be parsed, got cpus=%d, mem=%d, disk=%d, port=%d", cfg.CPUs, cfg.MemoryGB, cfg.DiskGB, cfg.SpicePort)
	}

	baseDiskPath := filepath.Join(tempHome, "disks", "e2e-vm.qcow2")
	if _, err := os.Stat(baseDiskPath); err != nil {
		t.Fatalf("Base disk missing at %s: %v", baseDiskPath, err)
	}

	overlayDiskPath := filepath.Join(tempHome, "disks", "e2e-vm-overlay.qcow2")
	if _, err := os.Stat(overlayDiskPath); !os.IsNotExist(err) {
		t.Fatalf("Overlay disk should NOT exist before commit")
	}

	// Test update command: remove ISO with positional name first
	if err := runUpdate([]string{"e2e-vm", "-remove-iso"}); err != nil {
		t.Fatalf("update command failed to remove ISO: %v", err)
	}

	// Test update command: attach ISO with positional name first
	if err := runUpdate([]string{"e2e-vm", "-iso", fakeISO}); err != nil {
		t.Fatalf("update command failed to attach ISO: %v", err)
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

	// Test subsequent commit with -remove-iso (positional name first)
	if err := runCommit([]string{"e2e-vm", "-remove-iso"}); err != nil {
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

	// Test delete command (positional name first)
	if err := runDelete([]string{"e2e-vm", "-force"}); err != nil {
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

func TestParseAll(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		expectedPos  []string
		expectedFlag string
		expectedNum  int
		expectedBool bool
		expectErr    bool
	}{
		{
			name:         "positional first, flags after",
			args:         []string{"myvm", "-iso", "/tmp/ubuntu.iso", "-cpus", "8", "-force"},
			expectedPos:  []string{"myvm"},
			expectedFlag: "/tmp/ubuntu.iso",
			expectedNum:  8,
			expectedBool: true,
		},
		{
			name:         "flags first, positional after",
			args:         []string{"-iso", "/tmp/ubuntu.iso", "-cpus", "8", "-force", "myvm"},
			expectedPos:  []string{"myvm"},
			expectedFlag: "/tmp/ubuntu.iso",
			expectedNum:  8,
			expectedBool: true,
		},
		{
			name:         "interspersed flags and multiple positional",
			args:         []string{"pos1", "-iso", "/tmp/test.iso", "pos2", "-cpus", "4", "pos3"},
			expectedPos:  []string{"pos1", "pos2", "pos3"},
			expectedFlag: "/tmp/test.iso",
			expectedNum:  4,
		},
		{
			name:         "terminator --",
			args:         []string{"pos1", "--", "-not-a-flag", "pos2"},
			expectedPos:  []string{"pos1", "-not-a-flag", "pos2"},
			expectedNum:  2,
		},
		{
			name:      "unknown flag returns error",
			args:      []string{"pos1", "-unknown", "pos2"},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			iso := fs.String("iso", "", "")
			cpus := fs.Int("cpus", 2, "")
			force := fs.Bool("force", false, "")

			pos, err := ParseAll(fs, tc.args)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(pos) != len(tc.expectedPos) {
				t.Fatalf("positional length mismatch: expected %v, got %v", tc.expectedPos, pos)
			}
			for i := range pos {
				if pos[i] != tc.expectedPos[i] {
					t.Errorf("pos[%d]: expected %s, got %s", i, tc.expectedPos[i], pos[i])
				}
			}
			if *iso != tc.expectedFlag {
				t.Errorf("iso flag: expected %q, got %q", tc.expectedFlag, *iso)
			}
			if *cpus != tc.expectedNum {
				t.Errorf("cpus flag: expected %d, got %d", tc.expectedNum, *cpus)
			}
			if *force != tc.expectedBool {
				t.Errorf("force flag: expected %v, got %v", tc.expectedBool, *force)
			}
		})
	}
}
