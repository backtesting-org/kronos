package cmd

import (
	backtesting "github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type AnalyzeCommandResult struct {
	fx.Out
	AnalyzeCommand *cobra.Command `name:"analyze"`
}

// NewAnalyzeCommand creates the analyze command
func NewAnalyzeCommand(handler backtesting.AnalyzeHandler) AnalyzeCommandResult {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze backtest results",
		RunE:  handler.Handle,
	}

	cmd.Flags().String("path", "./results", "Path to results directory")

	return AnalyzeCommandResult{
		AnalyzeCommand: cmd,
	}
}
