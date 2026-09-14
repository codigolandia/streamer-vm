package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"streamer-vm/internal/i18n"
)

// Command represents a CLI subcommand.
type Command struct {
	Name     string
	Short    string
	Long     string
	ShortKey string
	LongKey  string
	Run      func(args []string) error
}

func (c *Command) GetShort() string {
	if c.ShortKey != "" {
		return i18n.T(c.ShortKey)
	}
	return c.Short
}

func (c *Command) GetLong() string {
	if c.LongKey != "" {
		return i18n.T(c.LongKey)
	}
	return c.Long
}

var commands = map[string]*Command{}

// RegisterCommand registers a new subcommand.
func RegisterCommand(cmd *Command) {
	commands[cmd.Name] = cmd
}

// ParseAll parses all arguments provided with the given FlagSet,
// allowing flags to be placed anywhere (before, interspersed, or after positional arguments).
// It returns the remaining positional arguments in the order they were provided.
func ParseAll(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string

	for len(args) > 0 {
		if args[0] == "--" {
			positional = append(positional, args[1:]...)
			break
		}

		if err := fs.Parse(args); err != nil {
			return nil, err
		}

		rem := fs.Args()
		if len(rem) == 0 {
			break
		}

		positional = append(positional, rem[0])
		args = rem[1:]
	}

	return positional, nil
}

// extractLang filters out --lang / -lang options from CLI arguments.
func extractLang(args []string) (string, []string) {
	var lang string
	var filtered []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--lang" || arg == "-lang" {
			if i+1 < len(args) {
				lang = args[i+1]
				i++
				continue
			}
		} else if strings.HasPrefix(arg, "--lang=") {
			lang = strings.TrimPrefix(arg, "--lang=")
			continue
		} else if strings.HasPrefix(arg, "-lang=") {
			lang = strings.TrimPrefix(arg, "-lang=")
			continue
		}
		filtered = append(filtered, arg)
	}
	return lang, filtered
}

// Execute runs the CLI application with standard library argument parsing.
func Execute() {
	lang, args := extractLang(os.Args[1:])
	i18n.Init(lang)

	if len(args) == 0 {
		PrintHelp()
		os.Exit(0)
	}

	subCmd := args[0]
	if subCmd == "-v" || subCmd == "--version" {
		subCmd = "version"
	}
	if subCmd == "help" || subCmd == "-h" || subCmd == "--help" {
		if len(args) > 1 {
			if cmd, ok := commands[args[1]]; ok {
				fmt.Print(i18n.T("cli.usage_of", cmd.Name, cmd.GetLong()))
				_ = cmd.Run([]string{"-h"})
				return
			}
		}
		PrintHelp()
		return
	}

	cmd, ok := commands[subCmd]
	if !ok {
		fmt.Fprint(os.Stderr, i18n.T("cli.unknown_command", subCmd))
		os.Exit(1)
	}

	if err := cmd.Run(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// CommandGroup groups subcommands by lifecycle stage.
type CommandGroup struct {
	TitleKey string
	Names    []string
}

var commandWorkflow = []CommandGroup{
	{
		TitleKey: "cli.group.setup",
		Names:    []string{"init", "create", "update", "commit"},
	},
	{
		TitleKey: "cli.group.execution",
		Names:    []string{"start", "stop", "status", "spice-url"},
	},
	{
		TitleKey: "cli.group.maintenance",
		Names:    []string{"reset", "list", "delete"},
	},
}

// PrintHelp displays general CLI usage, the recommended workflow, and commands organized by lifecycle stage.
func PrintHelp() {
	fmt.Println(i18n.T("cli.description"))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage"))
	fmt.Printf("  %s\n", i18n.T("cli.usage_line"))
	fmt.Println()
	fmt.Println(i18n.T("cli.workflow_title"))
	fmt.Println("  1. streamer-vm init")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step1"))
	fmt.Println("  2. streamer-vm create <name> -iso <path>")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step2"))
	fmt.Println("  3. streamer-vm start <name> [--gui]")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step3"))
	fmt.Println("  4. streamer-vm stop <name>")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step4"))
	fmt.Println("  5. streamer-vm commit <name> --remove-iso")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step5"))
	fmt.Println("  6. streamer-vm start <name>")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step6"))
	fmt.Println("  7. streamer-vm reset <name>")
	fmt.Printf("     %s\n", i18n.T("cli.workflow.step7"))
	fmt.Println()
	fmt.Println(i18n.T("cli.commands_stage"))

	seen := make(map[string]bool)
	for _, group := range commandWorkflow {
		fmt.Printf("\n  %s\n", i18n.T(group.TitleKey))
		for _, name := range group.Names {
			if c, ok := commands[name]; ok {
				seen[name] = true
				fmt.Printf("    %-12s %s\n", c.Name, c.GetShort())
			}
		}
	}

	// Any unclassified commands
	var other []string
	for k := range commands {
		if !seen[k] {
			other = append(other, k)
		}
	}
	if len(other) > 0 {
		sort.Strings(other)
		fmt.Printf("\n  %s\n", i18n.T("cli.group.other"))
		for _, k := range other {
			c := commands[k]
			fmt.Printf("    %-12s %s\n", c.Name, c.GetShort())
		}
	}

	fmt.Println()
	fmt.Println(i18n.T("cli.flags"))
	fmt.Printf("  -h, --help    %s\n", i18n.T("cli.flags_help"))
	fmt.Printf("  -v, --version %s\n", i18n.T("cli.flags_version"))
	fmt.Printf("  --lang        %s\n", i18n.T("cli.flags_lang"))
}
