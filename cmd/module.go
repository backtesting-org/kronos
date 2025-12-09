package cmd

import (
	"go.uber.org/fx"
)

// Module provides all command-related dependencies
var Module = fx.Module("commands",
	fx.Provide(
		NewRootCommand,
		NewInitCommand,
		//NewLiveCommand,
		NewBacktestCommand,
		NewAnalyzeCommand,
		NewVersionCommand,
		NewRunStrategyCommand,
		NewCommands,
	),
	fx.Invoke(registerCommands),
)

type registerCommandsParams struct {
	fx.In

	Root        *RootCommand
	Cmds        *Commands
	RunStrategy *RunStrategyCommand
}

// registerCommands wires up the command tree
func registerCommands(p registerCommandsParams) {
	p.Root.Cmd.AddCommand(p.Cmds.Init)
	//p.Root.Cmd.AddCommand(p.Cmds.Live)
	p.Root.Cmd.AddCommand(p.Cmds.Backtest)
	p.Root.Cmd.AddCommand(p.Cmds.Analyze)
	p.Root.Cmd.AddCommand(p.Cmds.Version)
	p.Root.Cmd.AddCommand(p.RunStrategy.Cmd)
}
