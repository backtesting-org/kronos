package browse

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type StrategyListView interface {
	tea.Model
}

type strategyListView struct {
	strategies      []strategy.Strategy
	cursor          int
	pageSize        int
	pageNum         int
	compileService  shared.CompileService
	strategyService strategy.StrategyConfig
	detailFactory   StrategyDetailViewFactory
}

// newStrategyListView is the private constructor called by the factory
func newStrategyListView(
	compileService shared.CompileService,
	strategyService strategy.StrategyConfig,
	detailFactory StrategyDetailViewFactory,
) tea.Model {
	view := &strategyListView{
		compileService:  compileService,
		strategyService: strategyService,
		detailFactory:   detailFactory,
	}

	view.strategies, _ = strategyService.FindStrategies()
	view.pageSize = 5
	view.pageNum = 1

	return view
}

func (m *strategyListView) Init() tea.Cmd {
	return nil
}

func (m *strategyListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			return m, bubblon.Cmd(bubblon.Close())
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.strategies)-1 {
				m.cursor++
			}
		case "enter":
			// Create detail view with selected strategy using factory
			selectedStrat := &m.strategies[m.cursor]
			detailView := m.detailFactory(selectedStrat)

			// Use Bubblon to push the new view onto the stack
			return m, bubblon.Open(detailView)
		}
	}
	return m, nil
}

func (m *strategyListView) View() string {
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
