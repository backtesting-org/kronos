package cmd

import (
	backtesting "github.com/backtesting-org/kronos-cli/internal/backtesting/types"
	core "github.com/backtesting-org/kronos-cli/internal/handlers"
	live "github.com/backtesting-org/kronos-cli/internal/live/types"
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	"github.com/spf13/cobra"
)

// Commands struct holds all command handlers
type Commands struct {
	Root     *cobra.Command
	Init     *cobra.Command
	Live     *cobra.Command
	Backtest *cobra.Command
	Analyze  *cobra.Command
	Version  *cobra.Command
}

// NewCommands creates all cobra commands and wires them to handlers
func NewCommands(
	rootHandler core.RootHandler,
	initHandler setup.InitHandler,
	liveHandler live.LiveHandler,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
) *Commands {
	// Root command
	rootCmd := &cobra.Command{
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
		RunE: rootHandler.Handle,
	}

	rootCmd.PersistentFlags().Bool("cli", false, "Use CLI mode instead of interactive TUI")

	// Init command
	initCmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new Kronos project",
		RunE:  initHandler.Handle,
	}

	// Live command
	liveCmd := &cobra.Command{
		Use:   "live",
		Short: "Deploy strategies to live trading",
		RunE:  liveHandler.Handle,
	}
	liveCmd.Flags().String("strategy", "", "Strategy name for non-interactive mode")
	liveCmd.Flags().String("exchange", "", "Connectors for non-interactive mode")

	// Backtest command
	backtestCmd := &cobra.Command{
		Use:   "backtest",
		Short: "Run backtests",
		RunE:  backtestHandler.Handle,
	}
	backtestCmd.Flags().String("config", "", "Path to backtest config file (for CLI mode)")

	// Analyze command
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze backtest results",
		RunE:  analyzeHandler.Handle,
	}
	analyzeCmd.Flags().String("path", "./results", "Path to results directory")

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Kronos CLI v0.1.0")
		},
	}

	// Add subcommands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(liveCmd)
	rootCmd.AddCommand(backtestCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(versionCmd)

	return &Commands{
		Root:     rootCmd,
		Init:     initCmd,
		Live:     liveCmd,
		Backtest: backtestCmd,
		Analyze:  analyzeCmd,
		Version:  versionCmd,
	}
}
