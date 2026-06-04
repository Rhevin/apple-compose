package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func unsupportedCmd(use, short, reason string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%q is not supported: %s", use, reason)
		},
	}
}

var (
	pauseCmd = unsupportedCmd(
		"pause [service...]",
		"Pause services (not supported)",
		"Apple container CLI has no pause/unpause capability",
	)
	unpauseCmd = unsupportedCmd(
		"unpause [service...]",
		"Unpause services (not supported)",
		"Apple container CLI has no pause/unpause capability",
	)
	eventsCmd = unsupportedCmd(
		"events [service...]",
		"Receive real-time events (not supported)",
		"Apple container CLI has no real-time event stream",
	)
	waitCmd = unsupportedCmd(
		"wait <service...>",
		"Block until containers stop (not supported)",
		"Apple container CLI has no 'container wait' command",
	)
	scaleCmd = unsupportedCmd(
		"scale <service=N...>",
		"Scale services (not supported)",
		"multiple replicas are not yet supported",
	)
)
