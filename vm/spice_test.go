package vm

import (
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestGetSpiceURL(t *testing.T) {
	urlUnix := GetSpiceURL("127.0.0.1", 5900, "/tmp/test.sock")
	if urlUnix != "spice+unix:///tmp/test.sock" {
		t.Errorf("Expected spice+unix:///tmp/test.sock, got %s", urlUnix)
	}

	urlTCP := GetSpiceURL("127.0.0.1", 5900, "")
	if urlTCP != "spice://127.0.0.1:5900" {
		t.Errorf("Expected spice://127.0.0.1:5900, got %s", urlTCP)
	}
}

func TestIsSpiceReadyAndWait(t *testing.T) {
	tempDir := t.TempDir()
	sockPath := filepath.Join(tempDir, "test.sock")

	// Not ready
	if IsSpiceReady(sockPath, "", 0, 50*time.Millisecond) {
		t.Errorf("Expected false for closed unix socket")
	}

	// Start unix listener
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to open unix test listener: %v", err)
	}
	defer l.Close()

	if !IsSpiceReady(sockPath, "", 0, 200*time.Millisecond) {
		t.Errorf("Expected true for listening unix socket %s", sockPath)
	}

	if !WaitUntilSpiceReady(sockPath, "", 0, 1*time.Second) {
		t.Errorf("WaitUntilSpiceReady timed out for active unix listener")
	}
}
