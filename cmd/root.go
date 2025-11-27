package cmd

import (
	core "github.com/backtesting-org/kronos-cli/internal/handlers"
	"github.com/spf13/cobra"
)

// RootCommand wraps the root cobra command
type RootCommand struct {
	Cmd *cobra.Command
}

// NewRootCommand creates the root command
func NewRootCommand(handler core.RootHandler) *RootCommand {
	cmd := &cobra.Command{
		Use:   "kronos",
		Short: "Kronos - Trading infrastructure platform",
		Long: `Kronos CLI - Backtesting and live trading infrastructure

Use Kronos to:
  • Configure backtests via YAML
  • Run backtests locally with deterministic simulation
  • Deploy strategies live
  • Analyze results

Examples:
  kronos                         Launch interactive TUI menu (default)
  kronos --cli                   Show traditional CLI help
  kronos init my-project         Create a new project (CLI mode)
  kronos init                    Create a new project (TUI mode)
  kronos backtest --cli --config backtest.yaml    Run backtest via CLI
  kronos backtest                Run backtest via TUI
  kronos live --cli --strategy arbitrage --exchange binance    Run live via CLI
  kronos live                    Run live via TUI`,
		RunE: handler.Handle,
	}

	cmd.PersistentFlags().Bool("cli", false, "Use CLI mode instead of interactive TUI")

	return &RootCommand{Cmd: cmd}
}
