package cmd

import (
	"os"

	"go.uber.org/fx"
)

// Module provides all command-related dependencies
var Module = func() fx.Option {
	providers := []interface{}{
		NewRootCommand,
		NewInitCommand,
		NewLiveCommand,
		NewBacktestCommand,
		NewAnalyzeCommand,
		NewVersionCommand,
		NewCommands,
	}

	// Only provide RunStrategyCommand when running run-strategy
	if len(os.Args) > 1 && os.Args[1] == "run-strategy" {
		providers = append(providers, NewRunStrategyCommand)
	}

	return fx.Module("commands",
		fx.Provide(providers...),
		fx.Invoke(registerCommands),
	)
}()

type registerCommandsParams struct {
	fx.In

	Root        *RootCommand
	Cmds        *Commands
	RunStrategy *RunStrategyCommand `optional:"true"`
}

// registerCommands wires up the command tree
func registerCommands(p registerCommandsParams) {
	p.Root.Cmd.AddCommand(p.Cmds.Init)
	p.Root.Cmd.AddCommand(p.Cmds.Live)
	p.Root.Cmd.AddCommand(p.Cmds.Backtest)
	p.Root.Cmd.AddCommand(p.Cmds.Analyze)
	p.Root.Cmd.AddCommand(p.Cmds.Version)

	// Only add run-strategy if it was provided
	if p.RunStrategy != nil {
		p.Root.Cmd.AddCommand(p.RunStrategy.Cmd)
	}
}
