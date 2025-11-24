package handlers

import (
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
		// Run interactive TUI flow
		strategyExample, projectName, err := RunInitTUI()
		if err != nil {
			return err
		}
		return h.scaffoldService.CreateProjectWithStrategy(projectName, strategyExample)
	}

	name := args[0]
	return h.scaffoldService.CreateProject(name)
}

func (h *InitHandler) HandleWithStrategy(strategyExample, name string) error {
	return h.scaffoldService.CreateProjectWithStrategy(name, strategyExample)
}
