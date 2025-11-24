package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// LiveHandler handles the live command
type LiveHandler struct {
	liveService *services.LiveService
}

func NewLiveHandler(liveService *services.LiveService) *LiveHandler {
	return &LiveHandler{
		liveService: liveService,
	}
}

func (h *LiveHandler) Handle(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// TODO: Handle non-interactive mode with flags
		return nil
	}

	return h.liveService.RunSelectionTUI()
}
