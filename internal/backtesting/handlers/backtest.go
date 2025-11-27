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
	// Check if we're in CLI mode
	cliMode, _ := cmd.Root().PersistentFlags().GetBool("cli")
	configPath, _ := cmd.Flags().GetString("config")

	// Use CLI mode if --cli flag is set OR if --config is provided
	if cliMode || configPath != "" {
		if configPath == "" {
			return fmt.Errorf("CLI mode requires --config flag\n\nUsage:\n  kronos backtest --cli --config <path>")
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

	// Default to TUI mode (interactive)
	return h.backtestService.RunInteractive()
}
