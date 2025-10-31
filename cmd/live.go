package cmd

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Start live trading",
	Long: `Start live trading with your strategy.
	
Examples:
  # Start live trading with config
  kronos live
  
  # Dry-run mode (paper trading)
  kronos live --dry-run
  
  # Start with custom config
  kronos live --config production.yml`,
	RunE: runLive,
}

var liveDryRun bool

func init() {
	liveCmd.Flags().StringVar(&configPath, "config", "kronos.yml", "Path to config file")
	liveCmd.Flags().BoolVar(&liveDryRun, "dry-run", false, "Paper trading mode (no real orders)")
}

func runLive(cmd *cobra.Command, args []string) error {
	ui.ShowBanner()
	ui.Warning("Live trading feature coming soon!")
	fmt.Println()
	
	ui.Info("This will include:")
	fmt.Println("  • Real-time strategy execution")
	fmt.Println("  • Exchange API integration")
	fmt.Println("  • Risk management controls")
	fmt.Println("  • Real-time P&L tracking")
	fmt.Println("  • Paper trading mode")
	fmt.Println()
	
	ui.Warning("⚠️  Live trading involves real money. Use with caution!")
	
	return nil
}

