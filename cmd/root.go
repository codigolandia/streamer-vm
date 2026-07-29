package cmd

import (
	"fmt"
	"os"
	"sort"
)

// Command represents a CLI subcommand.
type Command struct {
	Name  string
	Short string
	Long  string
	Run   func(args []string) error
}

var commands = map[string]*Command{}

// RegisterCommand registers a new subcommand.
func RegisterCommand(cmd *Command) {
	commands[cmd.Name] = cmd
}

// Execute runs the CLI application with standard library argument parsing.
func Execute() {
	args := os.Args[1:]
	if len(args) == 0 {
		PrintHelp()
		os.Exit(0)
	}

	subCmd := args[0]
	if subCmd == "help" || subCmd == "-h" || subCmd == "--help" {
		if len(args) > 1 {
			if cmd, ok := commands[args[1]]; ok {
				fmt.Printf("Usage of streamer-vm %s:\n\n%s\n\n", cmd.Name, cmd.Long)
				_ = cmd.Run([]string{"-h"})
				return
			}
		}
		PrintHelp()
		return
	}

	cmd, ok := commands[subCmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'streamer-vm help' for usage.\n", subCmd)
		os.Exit(1)
	}

	if err := cmd.Run(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// PrintHelp displays general CLI usage and available subcommands.
func PrintHelp() {
	fmt.Println("streamer-vm is a CLI tool designed to manage QEMU+KVM virtual machines with VirGL")
	fmt.Println("GPU acceleration, SPICE display, and seamless OBS Studio capture integration.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  streamer-vm <command> [flags] [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")

	var keys []string
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		c := commands[k]
		fmt.Printf("  %-12s %s\n", c.Name, c.Short)
	}
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --help   Show help information")
}
