package cmd

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Start live trading",
	Long: `Start live trading with your strategy.
	
By default, opens an interactive TUI to select and deploy your strategy.
Use --non-interactive flag for direct execution.

Examples:
  # Interactive strategy selector (default)
  kronos live
  
  # Direct execution (for scripts/CI)
  kronos live --non-interactive --strategy momentum --exchange paradex
  
  # Paper trading mode
  kronos live --dry-run`,
	RunE: runLive,
}

var (
	liveStrategy string
	liveExchange string
	liveDryRun   bool
)

func init() {
	liveCmd.Flags().StringVar(&liveStrategy, "strategy", "", "Specific strategy to deploy (requires --non-interactive)")
	liveCmd.Flags().StringVar(&liveExchange, "exchange", "", "Exchange to use (requires --non-interactive)")
	liveCmd.Flags().BoolVar(&liveDryRun, "dry-run", false, "Paper trading mode (no real orders)")
}

func runLive(cmd *cobra.Command, args []string) error {
	// Check if non-interactive mode
	if nonInteractive {
		// Validate required flags
		if liveStrategy == "" || liveExchange == "" {
			return fmt.Errorf("--strategy and --exchange are required in non-interactive mode")
		}

		// Direct execution
		fmt.Printf("🚀 Starting %s strategy on %s...\n", liveStrategy, liveExchange)
		if liveDryRun {
			fmt.Println("📝 Mode: PAPER TRADING")
		} else {
			fmt.Println("🔴 Mode: LIVE TRADING")
		}

		// TODO: Execute live-trading binary with params
		fmt.Printf("\nCommand: live-trading --exchange %s --strategy ./strategies/%s/strategy.so\n",
			liveExchange, liveStrategy)

		return nil
	}

	// Launch the interactive TUI
	if err := live.RunSelectionTUI(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}
