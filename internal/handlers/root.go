package handlers

import (
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	"github.com/backtesting-org/kronos-cli/internal/strategies"
	backtesting "github.com/backtesting-org/kronos-cli/internal/strategies/backtest/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type RootHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// RootHandler handles the root command and main menu
type rootHandler struct {
	strategyBrowser strategies.StrategyBrowser
	initHandler     setup.InitHandler
	backtestHandler backtesting.BacktestHandler
	analyzeHandler  backtesting.AnalyzeHandler
}

func NewRootHandler(
	strategyBrowser strategies.StrategyBrowser,
	initHandler setup.InitHandler,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
) RootHandler {
	return &rootHandler{
		strategyBrowser: strategyBrowser,
		initHandler:     initHandler,
		backtestHandler: backtestHandler,
		analyzeHandler:  analyzeHandler,
	}
}

func (h *rootHandler) Handle(cmd *cobra.Command, args []string) error {
	cliMode, _ := cmd.Flags().GetBool("cli")

	if cliMode || len(args) > 0 {
		return cmd.Help()
	}

	return h.runMainMenu(cmd)
}

func (h *rootHandler) runMainMenu(rootCmd *cobra.Command) error {
	m := mainMenuModel{
		choices: []string{
			"Strategies",
			"Settings",
			"Help",
			"Create New Project",
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
	case "Strategies":
		return h.strategyBrowser.Handle(rootCmd, []string{})
	case "Settings":
		return h.handleSettings(rootCmd)
	case "Help":
		return showHelp()
	case "Create New Project":
		return h.handleCreateProject(rootCmd)
	}

	return nil
}

// handleSettings opens the settings/configuration menu
func (h *rootHandler) handleSettings(_ *cobra.Command) error {
	// For now, this is a placeholder that will open a settings TUI
	// TODO: Implement settings UI to edit exchanges/connectors
	return nil
}
