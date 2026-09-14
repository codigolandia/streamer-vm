package cmd

import (
	"flag"
	"fmt"

	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "spice-url",
		ShortKey: "cmd.spice_url.short",
		LongKey:  "cmd.spice_url.long",
		Short:    "Print the SPICE connection URL for a VM",
		Long:     "Displays the spice:// or spice+unix:// connection URI to connect using spicy or virt-viewer.",
		Run:      runSpiceURL,
	})
}

func runSpiceURL(args []string) error {
	fs := flag.NewFlagSet("spice-url", flag.ContinueOnError)
	positional, err := ParseAll(fs, args)
	if err != nil {
		return err
	}

	if len(positional) < 1 {
		return fmt.Errorf("VM name argument is required: streamer-vm spice-url <name>")
	}

	name := positional[0]
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
