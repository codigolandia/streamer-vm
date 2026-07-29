package vm

import (
	"net"
	"testing"
	"time"
)

func TestGetSpiceURL(t *testing.T) {
	url := GetSpiceURL("127.0.0.1", 5900)
	if url != "spice://127.0.0.1:5900" {
		t.Errorf("Expected spice://127.0.0.1:5900, got %s", url)
	}

	urlDefaultHost := GetSpiceURL("", 5905)
	if urlDefaultHost != "spice://127.0.0.1:5905" {
		t.Errorf("Expected spice://127.0.0.1:5905, got %s", urlDefaultHost)
	}
}

func TestAllocateFreeSpicePort(t *testing.T) {
	port, err := AllocateFreeSpicePort(5900, 5999)
	if err != nil {
		t.Fatalf("AllocateFreeSpicePort failed: %v", err)
	}
	if port < 5900 || port > 5999 {
		t.Errorf("Allocated port %d out of bounds", port)
	}
}

func TestIsSpicePortListeningAndWait(t *testing.T) {
	// Not listening port
	if IsSpicePortListening("127.0.0.1", 59876, 50*time.Millisecond) {
		t.Errorf("Expected false for closed port")
	}

	// Start temporary listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to open test listener: %v", err)
	}
	defer l.Close()

	addr := l.Addr().(*net.TCPAddr)
	port := addr.Port

	if !IsSpicePortListening("127.0.0.1", port, 200*time.Millisecond) {
		t.Errorf("Expected true for listening port %d", port)
	}

	if !WaitUntilSpiceReady("127.0.0.1", port, 1*time.Second) {
		t.Errorf("WaitUntilSpiceReady timed out for active listener on port %d", port)
	}
}
