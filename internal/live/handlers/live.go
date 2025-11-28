package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/handlers/cli"
	"github.com/backtesting-org/kronos-cli/internal/live/handlers/interactive"
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/spf13/cobra"
)

// unifiedLiveHandler routes between CLI and TUI modes
type unifiedLiveHandler struct {
	cliHandler cli.CLILiveHandler
	tuiHandler interactive.LiveInteractive
}

func NewLiveHandler(
	cliHandler cli.CLILiveHandler,
	tuiHandler interactive.LiveInteractive,
) types.LiveHandler {
	return &unifiedLiveHandler{
		cliHandler: cliHandler,
		tuiHandler: tuiHandler,
	}
}

func (h *unifiedLiveHandler) Handle(cmd *cobra.Command, args []string) error {
	// Check if we're in CLI mode
	cliMode, _ := cmd.Root().PersistentFlags().GetBool("cli")

	// Check if CLI-specific flags are provided
	strategy, _ := cmd.Flags().GetString("strategy")
	exchange, _ := cmd.Flags().GetString("exchange")

	// Use CLI mode if --cli flag is set OR if CLI-specific flags are provided
	if cliMode || strategy != "" || exchange != "" {
		if strategy == "" || exchange == "" {
			return fmt.Errorf("CLI mode requires both --strategy and --exchange flags\n\nUsage:\n  kronos live --cli --strategy <name> --exchange <name>")
		}
		return h.cliHandler.Handle(cmd, args)
	}

	// Default to TUI mode
	return h.tuiHandler.Run()
}
