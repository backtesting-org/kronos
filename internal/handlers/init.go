package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// InitHandler handles the init command
type InitHandler struct {
	scaffoldService *services.ScaffoldService
}

func NewInitHandler(scaffoldService *services.ScaffoldService) *InitHandler {
	return &InitHandler{
		scaffoldService: scaffoldService,
	}
}

func (h *InitHandler) Handle(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name required")
	}

	name := args[0]
	return h.scaffoldService.CreateProject(name)
}
