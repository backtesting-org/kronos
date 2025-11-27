package cli

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/live-trading/pkg/runtime"
	"github.com/spf13/cobra"
)

type CLILiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// liveHandler handles the live command in CLI mode
type liveHandler struct {
	runtime        runtime.Runtime
	strategyConfig strategy.StrategyConfig
	liveService    types.LiveService
}

func NewCLILiveHandler(
	liveService types.LiveService,
	strategyConfig strategy.StrategyConfig,
	runtime runtime.Runtime,
) CLILiveHandler {
	return &liveHandler{
		liveService:    liveService,
		strategyConfig: strategyConfig,
		runtime:        runtime,
	}
}

func (h *liveHandler) Handle(
	cmd *cobra.Command,
	args []string,
) error {
	strategyName, _ := cmd.Flags().GetString("strategy")
	exchange, _ := cmd.Flags().GetString("exchange")

	fmt.Printf("🚀 Starting live trading in CLI mode\n")
	fmt.Printf("   Strategy: %s\n", strategyName)
	fmt.Printf("   Exchange: %s\n\n", exchange)

	strategies, err := h.strategyConfig.FindStrategies()
	if err != nil {
		return err
	}

	// Find the requested strategy
	var selectedStrategy *strategy.Strategy
	for i := range strategies {
		if strategies[i].Name == strategyName {
			selectedStrategy = &strategies[i]
			break
		}
	}

	if selectedStrategy == nil {
		return fmt.Errorf("strategy '%s' not found", strategyName)
	}

	fmt.Printf("✓ Found strategy: %s\n", selectedStrategy.Name)

	// TODO: Load exchange config and execute
	// This is where you'd validate credentials and start the strategy

	return nil
}
