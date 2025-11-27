package handlers

import (
	"fmt"

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
