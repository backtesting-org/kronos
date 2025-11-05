package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "kronos",
	Short: "Kronos - Trading infrastructure platform",
	Long: `Kronos CLI - Beautiful backtesting and live trading infrastructure

Use Kronos to:
  • Configure backtests via YAML
  • Run backtests locally with deterministic simulation
  • Deploy strategies live
  • Analyze results

Examples:
  kronos init my-project         Create a new project
  kronos backtest                Run backtest
  kronos backtest --interactive  Interactive mode
  kronos analyze                 Analyze results`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Kronos CLI v%s\n", version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(backtestCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(liveCmd)
	rootCmd.AddCommand(versionCmd)
}
