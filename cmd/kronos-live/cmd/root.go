package cmd

import (
	"github.com/spf13/cobra"
)

type RootCommand struct {
	Cmd *cobra.Command
}

func NewRootCommand(liveCmd *LiveCommand) *RootCommand {
	root := &cobra.Command{
		Use:   "kronos-live",
		Short: "Kronos Live Trading Runtime",
		Long:  `The runtime binary for executing live trading strategies`,
	}

	root.AddCommand(liveCmd.Cmd)

	return &RootCommand{Cmd: root}
}
