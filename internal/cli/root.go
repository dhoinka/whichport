package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dhoinka/whichport/internal/ports"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var (
		port     int
		protocol string
		jsonOut  bool
		noColor  bool
	)

	cmd := &cobra.Command{
		Use:           "whichport",
		Short:         "List applications currently listening on ports",
		Long:          "whichport shows the local processes that currently listen on TCP or UDP ports, including the port, pid, command name, full command line, and executable path.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       "dev",
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := ports.NewQuery(port, protocol)
			if err != nil {
				return err
			}

			listeners, err := ports.Discover(cmd.Context(), query)
			if err != nil {
				return err
			}

			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(listeners)
			}

			renderer := NewRenderer(cmd.OutOrStdout(), !noColor)
			if len(listeners) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), renderer.EmptyState())
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), renderer.Render(listeners))
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "Filter to a specific port")
	cmd.Flags().StringVar(&protocol, "protocol", "all", "Protocol filter: tcp, udp, or all")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print listeners as JSON")
	cmd.Flags().BoolVar(&noColor, "no-color", !isTerminal(os.Stdout.Fd()), "Disable color output")

	cmd.SetVersionTemplate(strings.TrimSpace(`
{{printf "%s\n" .Version}}
`))

	return cmd
}
