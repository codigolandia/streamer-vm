package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"streamer-vm/internal"
	"streamer-vm/internal/i18n"
	"streamer-vm/vm"
)

func init() {
	RegisterCommand(&Command{
		Name:     "list",
		ShortKey: "cmd.list.short",
		LongKey:  "cmd.list.long",
		Short:    "List all virtual machines",
		Long:     "Lists all configured virtual machines and their current operational status.",
		Run:      runList,
	})
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	home := vm.GetStreamerHome()
	vms, err := vm.ListVMs(home)
	if err != nil {
		return err
	}

	if len(vms) == 0 {
		internal.Info(i18n.T("status.no_vms"))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, i18n.T("list.header"))

	for _, cfg := range vms {
		running := vm.IsVMRunning(home, cfg.Name)
		statusStr := i18n.T("state.stopped")
		pidStr := "-"
		if running {
			statusStr = i18n.T("state.running")
			pidStr = fmt.Sprintf("%d", vm.GetVMPID(home, cfg.Name))
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			cfg.Name,
			statusStr,
			pidStr,
			cfg.SpicePort,
			cfg.CPUs,
			cfg.MemoryGB,
			cfg.DiskGB,
			cfg.CreatedAt.Local().Format("2006-01-02 15:04"),
		)
	}
	w.Flush()

	return nil
}
