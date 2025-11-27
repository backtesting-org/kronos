package cmd

import (
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type InitCommandResult struct {
	fx.Out
	InitCommand *cobra.Command `name:"init"`
}

// NewInitCommand creates the init command
func NewInitCommand(handler setup.InitHandler) InitCommandResult {
	return InitCommandResult{
		InitCommand: &cobra.Command{
			Use:   "init <name>",
			Short: "Create a new Kronos project",
			RunE:  handler.Handle,
		},
	}
}
