package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies"
	backtesting "github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/monitor"
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type RootHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// RootHandler handles the root command and main menu
type rootHandler struct {
	strategyBrowser    strategies.StrategyBrowser
	initHandler        setup.InitHandler
	backtestHandler    backtesting.BacktestHandler
	analyzeHandler     backtesting.AnalyzeHandler
	monitorViewFactory monitor.MonitorViewFactory
}

func NewRootHandler(
	strategyBrowser strategies.StrategyBrowser,
	initHandler setup.InitHandler,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
	monitorViewFactory monitor.MonitorViewFactory,
) RootHandler {
	return &rootHandler{
		strategyBrowser:    strategyBrowser,
		initHandler:        initHandler,
		backtestHandler:    backtestHandler,
		analyzeHandler:     analyzeHandler,
		monitorViewFactory: monitorViewFactory,
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
			"Monitor",
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
	case "Monitor":
		return h.handleMonitor(rootCmd)
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

// handleMonitor opens the monitor TUI for running strategies
func (h *rootHandler) handleMonitor(_ *cobra.Command) error {
	monitorView := h.monitorViewFactory()
	p := tea.NewProgram(monitorView, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
