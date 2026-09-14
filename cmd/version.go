package cmd

import (
	"fmt"
)

var (
	// Version is the application version, injected at build time via ldflags.
	Version = "dev"
	// Commit is the git commit hash, injected at build time via ldflags.
	Commit = "none"
	// Date is the build date, injected at build time via ldflags.
	Date = "unknown"
)

// SetVersionInfo sets the build-time version metadata.
func SetVersionInfo(v, c, d string) {
	if v != "" {
		Version = v
	}
	if c != "" {
		Commit = c
	}
	if d != "" {
		Date = d
	}
}

func init() {
	RegisterCommand(&Command{
		Name:     "version",
		Short:    "Exibe a versão do streamer-vm e informações de compilação",
		ShortKey: "cmd.version.short",
		Run: func(args []string) error {
			fmt.Printf("streamer-vm %s (commit %s, built at %s)\n", Version, Commit, Date)
			return nil
		},
	})
}
