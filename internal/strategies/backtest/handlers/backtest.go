package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/services/compile"
	types2 "github.com/backtesting-org/kronos-cli/internal/strategies/backtest/types"
	"github.com/spf13/cobra"
)

// backtestHandler handles the backtest command
type backtestHandler struct {
	backtestService types2.BacktestService
	compileService  compile.CompileService
}

func NewBacktestHandler(backtestService types2.BacktestService, compileService compile.CompileService) types2.BacktestHandler {
	return &backtestHandler{
		backtestService: backtestService,
		compileService:  compileService,
	}
}

func (h *backtestHandler) Handle(cmd *cobra.Command, args []string) error {
	// Default to TUI mode (interactive)
	return h.backtestService.RunInteractive()
}
