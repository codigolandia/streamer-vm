package vm

import (
	"fmt"
	"net"
	"time"
)

// DefaultSpiceStartPort is the default starting port for Spice allocation (5900).
const DefaultSpiceStartPort = 5900

// DefaultSpiceEndPort is the upper limit for Spice allocation (5999).
const DefaultSpiceEndPort = 5999

// AllocateFreeSpicePort finds an available TCP port on 127.0.0.1 within [startPort, endPort].
func AllocateFreeSpicePort(startPort, endPort int) (int, error) {
	if startPort <= 0 {
		startPort = DefaultSpiceStartPort
	}
	if endPort <= 0 {
		endPort = DefaultSpiceEndPort
	}

	for port := startPort; port <= endPort; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			l.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free Spice ports available in range %d-%d", startPort, endPort)
}

// IsSpicePortListening checks if a connection can be established to the Spice server.
func IsSpicePortListening(host string, port int, timeout time.Duration) bool {
	if host == "" {
		host = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// WaitUntilSpiceReady polls the Spice server until it is listening or maxWait duration passes.
func WaitUntilSpiceReady(host string, port int, maxWait time.Duration) bool {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if IsSpicePortListening(host, port, 500*time.Millisecond) {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// GetSpiceURL returns the formatted Spice URL.
func GetSpiceURL(host string, port int) string {
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("spice://%s:%d", host, port)
}
