package cmd

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/scaffold"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Initialize a new project",
	Long: `Initialize a new project directory.

Creates a new project directory with the cash_carry example from the SDK.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("project name is required\n\nUsage: kronos init <project-name>\nExample: kronos init my-trading-bot")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected 1 project name")
		}
		return nil
	},
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Show banner
	ui.ShowBanner()

	projectName := args[0]
	scaffolder := scaffold.NewScaffolder()
	return scaffolder.CreateProject(projectName)
}
