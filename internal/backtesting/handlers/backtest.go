package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/backtesting/types"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/spf13/cobra"
)

// backtestHandler handles the backtest command
type backtestHandler struct {
	backtestService types.BacktestService
	compileService  shared.CompileService
}

func NewBacktestHandler(backtestService types.BacktestService, compileService shared.CompileService) types.BacktestHandler {
	return &backtestHandler{
		backtestService: backtestService,
		compileService:  compileService,
	}
}

func (h *backtestHandler) Handle(cmd *cobra.Command, args []string) error {
	// Default to TUI mode (interactive)
	return h.backtestService.RunInteractive()
}
