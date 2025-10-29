package cmd

import (
	"github.com/backtesting-org/kronos-cli/internal/scaffold"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Kronos project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := "my-kronos-project"
	if len(args) > 0 {
		projectName = args[0]
	}

	scaffolder := scaffold.NewScaffolder()
	return scaffolder.CreateProject(projectName)
}
