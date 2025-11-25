package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// BacktestHandler handles the backtest command
type BacktestHandler struct {
	backtestService *services.BacktestService
	compileService  *services.CompileService
}

func NewBacktestHandler(backtestService *services.BacktestService, compileService *services.CompileService) *BacktestHandler {
	return &BacktestHandler{
		backtestService: backtestService,
		compileService:  compileService,
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

	// Pre-compile all strategies before backtesting
	fmt.Println("🔍 Checking strategies...")
	h.compileService.PreCompileStrategies("./strategies")
	fmt.Println()

	return h.backtestService.ExecuteBacktest(cfg)
}
