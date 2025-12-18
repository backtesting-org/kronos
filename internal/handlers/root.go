package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/settings"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies"
	backtesting "github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/browse"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/monitor"
	"github.com/backtesting-org/kronos-cli/internal/router"
	setup "github.com/backtesting-org/kronos-cli/internal/setup/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type RootHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

// RootHandler handles the root command and main menu
type rootHandler struct {
	strategyBrowser      strategies.StrategyBrowser
	initHandler          setup.InitHandler
	backtestHandler      backtesting.BacktestHandler
	analyzeHandler       backtesting.AnalyzeHandler
	monitorViewFactory   monitor.MonitorViewFactory
	strategyListFactory  browse.StrategyListViewFactory
	settingsListFactory  settings.SettingsListViewFactory
	connectorFormFactory settings.ConnectorFormViewFactory
	deleteConfirmFactory settings.DeleteConfirmViewFactory
	router               router.Router
}

func NewRootHandler(
	strategyBrowser strategies.StrategyBrowser,
	initHandler setup.InitHandler,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
	monitorViewFactory monitor.MonitorViewFactory,
	strategyListFactory browse.StrategyListViewFactory,
	settingsListFactory settings.SettingsListViewFactory,
	connectorFormFactory settings.ConnectorFormViewFactory,
	deleteConfirmFactory settings.DeleteConfirmViewFactory,
	r router.Router,
) RootHandler {
	// Register ALL routes with the router at initialization
	r.RegisterRoute(router.RouteMonitor, func() tea.Model {
		return monitorViewFactory()
	})

	r.RegisterRoute(router.RouteStrategyList, func() tea.Model {
		return strategyListFactory()
	})

	r.RegisterRoute(router.RouteSettingsList, func() tea.Model {
		return settingsListFactory()
	})

	r.RegisterRoute(router.RouteSettingsCreate, func() tea.Model {
		return connectorFormFactory("", false)
	})

	r.RegisterRoute(router.RouteSettingsEdit, func() tea.Model {
		return connectorFormFactory("", true)
	})

	r.RegisterRoute(router.RouteSettingsDelete, func() tea.Model {
		return deleteConfirmFactory("")
	})

	return &rootHandler{
		strategyBrowser:      strategyBrowser,
		initHandler:          initHandler,
		backtestHandler:      backtestHandler,
		analyzeHandler:       analyzeHandler,
		monitorViewFactory:   monitorViewFactory,
		strategyListFactory:  strategyListFactory,
		settingsListFactory:  settingsListFactory,
		connectorFormFactory: connectorFormFactory,
		deleteConfirmFactory: deleteConfirmFactory,
		router:               r,
	}
}

func (h *rootHandler) Handle(cmd *cobra.Command, args []string) error {
	cliMode, _ := cmd.Flags().GetBool("cli")

	if cliMode || len(args) > 0 {
		return cmd.Help()
	}

	return h.runMainMenu(cmd)
}

func (h *rootHandler) runMainMenu(_ *cobra.Command) error {
	m := mainMenuModel{
		choices: []string{
			"Strategies",
			"Monitor",
			"Settings",
			"Help",
			"Create New Project",
		},
		router: h.router,
	}

	// Set main menu as the initial view in router
	h.router.SetInitialView(m)

	// Run the router ONCE - all navigation happens within this single program
	p := tea.NewProgram(h.router, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
