package cmd

import (
	"flag"
	"fmt"

	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:  "spice-url",
		Short: "Print the SPICE connection URL for a VM",
		Long:  "Outputs the spice:// URL (e.g. spice://127.0.0.1:5900) for connecting via GTK client or virt-viewer.",
		Run:   runSpiceURL,
	})
}

func runSpiceURL(args []string) error {
	fs := flag.NewFlagSet("spice-url", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm spice-url <name>")
	}

	name := fs.Arg(0)
	home := vm.GetStreamerHome()

	cfg, err := vm.LoadVMConfig(home, name)
	if err != nil {
		return err
	}

	spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort)
	fmt.Println(spiceURL)
	return nil
}
