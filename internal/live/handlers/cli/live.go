package cli

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/live-trading/pkg/runtime"
	"github.com/spf13/cobra"
)

// liveHandler handles the live command
type liveHandler struct {
	runtime     runtime.Runtime
	liveService types.LiveService
}

func NewLiveHandler(
	liveService types.LiveService,
	runtime runtime.Runtime,
) types.LiveHandler {
	return &liveHandler{
		liveService: liveService,
	}
}

func (h *liveHandler) Handle(
	cmd *cobra.Command,
	args []string,
) error {
	strategies, err := h.liveService.DiscoverStrategies()
	if err != nil {
		return err
	}

	for i, strategy := range strategies {
		fmt.Printf("Starting strategy %d/%d: %s\n", i+1, len(strategies), strategy.Name)
	}

	return nil
}
