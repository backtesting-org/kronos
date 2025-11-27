package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/setup/types"
	"github.com/spf13/cobra"
)

// InitHandler handles the init command
type initHandler struct {
	scaffoldService types.ScaffoldService
}

func NewInitHandler(scaffoldService types.ScaffoldService) types.InitHandler {
	return &initHandler{
		scaffoldService: scaffoldService,
	}
}

func (h *initHandler) Handle(cmd *cobra.Command, args []string) error {
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

func (h *initHandler) HandleWithStrategy(strategyExample, name string) error {
	return h.scaffoldService.CreateProjectWithStrategy(name, strategyExample)
}
