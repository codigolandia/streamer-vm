package guest

import (
	"testing"
)

func TestPrereqsCheck(t *testing.T) {
	// Execute CheckAllPrereqs and ensure report structure is valid
	report := CheckAllPrereqs()
	if len(report.Checks) == 0 {
		t.Errorf("Expected non-empty prerequisite checks report")
	}

	for _, check := range report.Checks {
		if check.Name == "" {
			t.Errorf("PrereqCheck has empty Name: %+v", check)
		}
	}

	// Test PrintReport output (does not crash)
	report.PrintReport()
}

func TestFindOVMFPaths(t *testing.T) {
	code, vars, err := FindOVMFPaths()
	if err != nil {
		t.Logf("FindOVMFPaths returned error on this host: %v", err)
	} else {
		if code == "" || vars == "" {
			t.Errorf("FindOVMFPaths returned empty paths: code=%s, vars=%s", code, vars)
		}
	}
}
