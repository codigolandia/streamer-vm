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
		Long:  "Outputs the SPICE URL (e.g. spice+unix:///path/to/socket) for connecting via GTK client or virt-viewer.",
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

	spiceSock := vm.GetSpiceSocketPath(home, name)
	spiceURL := vm.GetSpiceURL("127.0.0.1", cfg.SpicePort, spiceSock)
	fmt.Println(spiceURL)
	return nil
}
