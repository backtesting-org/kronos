package handlers

import (
	types2 "github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/backtesting-org/kronos-cli/pkg/strategy"
	"github.com/spf13/cobra"
)

// backtestHandler handles the backtest command
type backtestHandler struct {
	backtestService types2.BacktestService
	compileService  strategy.CompileService
}

func NewBacktestHandler(backtestService types2.BacktestService, compileService strategy.CompileService) types2.BacktestHandler {
	return &backtestHandler{
		backtestService: backtestService,
		compileService:  compileService,
	}
}

func (h *backtestHandler) Handle(cmd *cobra.Command, args []string) error {
	// Default to TUI mode (interactive)
	return h.backtestService.RunInteractive()
}
