package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// BacktestHandler handles the backtest command
type BacktestHandler struct {
	backtestService *services.BacktestService
}

func NewBacktestHandler(backtestService *services.BacktestService) *BacktestHandler {
	return &BacktestHandler{
		backtestService: backtestService,
	}
}

func (h *BacktestHandler) Handle(cmd *cobra.Command, args []string) error {
	interactiveMode, _ := cmd.Flags().GetBool("interactive")
	configPath, _ := cmd.Flags().GetString("config")

	if interactiveMode || configPath == "" {
		return h.backtestService.RunInteractive()
	}

	cfg, err := h.backtestService.LoadConfig(configPath)
	if err != nil {
		return err
	}

	// TODO: implement strategies discovery and compilation here

	return h.backtestService.ExecuteBacktest(cfg)
}
