package handlers

import (
	backtesting "github.com/backtesting-org/kronos-cli/internal/backtesting/types"
	liveTypes "github.com/backtesting-org/kronos-cli/internal/live/types"
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type RootHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// RootHandler handles the root command and main menu
type rootHandler struct {
	initHandler     setup.InitHandler
	liveHandler     liveTypes.LiveHandler
	backtestHandler backtesting.BacktestHandler
	analyzeHandler  backtesting.AnalyzeHandler
}

func NewRootHandler(
	initHandler setup.InitHandler,
	liveHandler liveTypes.LiveHandler,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
) RootHandler {
	return &rootHandler{
		initHandler:     initHandler,
		liveHandler:     liveHandler,
		backtestHandler: backtestHandler,
		analyzeHandler:  analyzeHandler,
	}
}

func (h *rootHandler) Handle(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive || len(args) > 0 {
		return cmd.Help()
	}

	return h.runMainMenu(cmd)
}

func (h *rootHandler) runMainMenu(rootCmd *cobra.Command) error {
	m := mainMenuModel{
		choices: []string{
			"Start Live Trading",
			"Run Backtest",
			"Analyze Results",
			"Create New Project",
			"Show Help",
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	result := finalModel.(mainMenuModel)
	if result.selected == "" {
		return nil
	}

	switch result.selected {
	case "Start Live Trading":
		return h.liveHandler.Handle(rootCmd, []string{})
	case "Run Backtest":
		return h.backtestHandler.Handle(rootCmd, []string{})
	case "Analyze Results":
		return h.analyzeHandler.Handle(rootCmd, []string{})
	case "Create New Project":
		return h.handleCreateProject(rootCmd)
	case "Show Help":
		return showHelp()
	}

	return nil
}
