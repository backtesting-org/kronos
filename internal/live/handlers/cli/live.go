package cli

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/live-trading/pkg/runtime"
	"github.com/spf13/cobra"
)

type CLILiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// liveHandler handles the live command in CLI mode
type liveHandler struct {
	runtime     runtime.Runtime
	liveService types.LiveService
}

func NewCLILiveHandler(
	liveService types.LiveService,
	runtime runtime.Runtime,
) CLILiveHandler {
	return &liveHandler{
		liveService: liveService,
		runtime:     runtime,
	}
}

func (h *liveHandler) Handle(
	cmd *cobra.Command,
	args []string,
) error {
	strategy, _ := cmd.Flags().GetString("strategy")
	exchange, _ := cmd.Flags().GetString("exchange")

	fmt.Printf("🚀 Starting live trading in CLI mode\n")
	fmt.Printf("   Strategy: %s\n", strategy)
	fmt.Printf("   Exchange: %s\n\n", exchange)

	strategies, err := h.liveService.DiscoverStrategies()
	if err != nil {
		return err
	}

	// Find the requested strategy
	var selectedStrategy *types.Strategy
	for i := range strategies {
		if strategies[i].Name == strategy {
			selectedStrategy = &strategies[i]
			break
		}
	}

	if selectedStrategy == nil {
		return fmt.Errorf("strategy '%s' not found", strategy)
	}

	fmt.Printf("✓ Found strategy: %s\n", selectedStrategy.Name)

	// TODO: Load exchange config and execute
	// This is where you'd validate credentials and start the strategy

	return nil
}
