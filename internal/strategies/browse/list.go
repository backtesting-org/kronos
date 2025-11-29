package browse

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
	tea "github.com/charmbracelet/bubbletea"
)

type StrategyListView interface {
	tea.Model
}

// strategyListView represents the strategy list view (STRATEGIES screen)
type strategyListView struct {
	strategies      []strategy.Strategy
	cursor          int
	pageSize        int
	pageNum         int
	router          router.Router
	compileService  shared.CompileService
	strategyService strategy.StrategyConfig
	detailView      StrategyDetailView
}

func NewStrategyListView(
	compileService shared.CompileService,
	strategyService strategy.StrategyConfig,
	detailView StrategyDetailView,
) StrategyListView {
	view := strategyListView{
		compileService:  compileService,
		strategyService: strategyService,
		detailView:      detailView,
	}

	view.strategies, _ = strategyService.FindStrategies()
	view.pageSize = 5
	view.pageNum = 1

	return view
}

func (m strategyListView) Init() tea.Cmd {
	return nil
}

func (m strategyListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.strategies)-1 {
				m.cursor++
			}
		case "enter":
			// Create detail view with services this view already has
			selectedStrat := &m.strategies[m.cursor]

			m.detailView.SetStrategy(selectedStrat)

			// Send navigation message with created view
			return m, func() tea.Msg {
				return router.NavigateMsg{
					Route: router.RouteStrategyDetail,
					View:  m.detailView,
				}
			}
		}
	case router.NavigateMsg:
		// Handle navigation
		switch msg.Route {
		case router.RouteMenu:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m strategyListView) View() string {
	if len(m.strategies) == 0 {
		return ui.TitleStyle.Render("STRATEGIES") + "\n\n" + ui.SubtitleStyle.Render("No strategies found. Create a new one to get started.")
	}

	var content string
	content += ui.TitleStyle.Render("STRATEGIES") + "\n"
	content += ui.SubtitleStyle.Render("Use arrow keys to navigate, Enter to select, q to quit") + "\n\n"

	// Display current page
	for i, strat := range m.strategies {
		exchanges := fmt.Sprintf("[%v]", strat.Exchanges)
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+strat.Name+" "+exchanges) + "\n"
		} else {
			content += ui.StrategyNameStyle.Render("  "+strat.Name+" "+exchanges) + "\n"
		}
	}

	// Show pagination info
	totalPages := (len(m.strategies) + m.pageSize - 1) / m.pageSize
	content += "\n" + ui.SubtitleStyle.Render(fmt.Sprintf("Page %d/%d", m.pageNum, totalPages))

	return ui.BoxStyle.Render(content)
}
