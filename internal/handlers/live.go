package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// LiveHandler handles the live command
type LiveHandler struct {
	liveService     *services.LiveService
	scaffoldService *services.ScaffoldService
}

func NewLiveHandler(liveService *services.LiveService, scaffoldService *services.ScaffoldService) *LiveHandler {
	return &LiveHandler{
		liveService:     liveService,
		scaffoldService: scaffoldService,
	}
}

func (h *LiveHandler) Handle(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// TODO: Handle non-interactive mode with flags
		return nil
	}

	err := h.liveService.RunSelectionTUI()

	// Check if user requested project initialization
	if err != nil && err.Error() == "INIT_PROJECT_REQUESTED" {
		fmt.Println("\n✨ Initializing new Kronos project...")

		// Use current directory name as project name
		projectName := "."

		if createErr := h.scaffoldService.CreateProject(projectName); createErr != nil {
			return fmt.Errorf("failed to initialize project: %w", createErr)
		}

		fmt.Println("\n✅ Project initialized successfully!")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Create strategies in ./strategies/ directory")
		fmt.Println("  2. Configure exchanges.yml with your credentials")
		fmt.Println("  3. Run 'kronos live' again to deploy")

		return nil
	}

	return err
}
