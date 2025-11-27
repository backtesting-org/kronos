package cmd

import (
	"go.uber.org/fx"
)

// Module provides all command-related dependencies
var Module = fx.Module("commands",
	fx.Provide(
		NewRootCommand,
		NewInitCommand,
		NewLiveCommand,
		NewBacktestCommand,
		NewAnalyzeCommand,
		NewVersionCommand,
		NewCommands,
	),
	fx.Invoke(registerCommands),
)

// registerCommands wires up the command tree
func registerCommands(root *RootCommand, cmds *Commands) {
	root.Cmd.AddCommand(cmds.Init)
	root.Cmd.AddCommand(cmds.Live)
	root.Cmd.AddCommand(cmds.Backtest)
	root.Cmd.AddCommand(cmds.Analyze)
	root.Cmd.AddCommand(cmds.Version)
}
