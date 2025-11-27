package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type VersionCommandResult struct {
	fx.Out
	VersionCommand *cobra.Command `name:"version"`
}

// NewVersionCommand creates the version command
func NewVersionCommand() VersionCommandResult {
	return VersionCommandResult{
		VersionCommand: &cobra.Command{
			Use:   "version",
			Short: "Show version information",
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Println("Kronos CLI v0.1.0")
			},
		},
	}
}
