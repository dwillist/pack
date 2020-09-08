package commands

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/buildpacks/pack/logging"
)

// Version shows the current pack version
func Version(logger logging.Logger, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: "Show current 'pack' version",
		RunE: LogError(logger, func(cmd *cobra.Command, args []string) error {
			logger.Info(strings.TrimSpace(version))
			return nil
		}),
	}
	AddHelpFlag(cmd, "version")
	return cmd
}
