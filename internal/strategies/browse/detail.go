package browse

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/strategies/compile"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
	tea "github.com/charmbracelet/bubbletea"
)

type StrategyDetailView interface {
	tea.Model
	SetStrategy(*strategy.Strategy)
}

// StrategyDetailView represents the strategy detail view with action options (STRATEGY screen)
type strategyDetailView struct {
	tea.Model
	strategy       *strategy.Strategy
	actions        []string
	cursor         int
	router         router.Router
	compileService shared.CompileService
}

// NewStrategyDetailView creates a strategy detail view with all dependencies
func NewStrategyDetailView(r router.Router, cs shared.CompileService) StrategyDetailView {
	return strategyDetailView{
		strategy:       nil,
		actions:        []string{"Compile", "Backtest", "Edit", "Delete"},
		cursor:         0,
		router:         r,
		compileService: cs,
	}
}

func (m strategyDetailView) SetStrategy(s *strategy.Strategy) {
	m.strategy = s
}

func (m strategyDetailView) Init() tea.Cmd {
	return nil
}

func (m strategyDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Navigate back to list
			return m, m.router.Navigate(router.RouteStrategyList)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
		case "enter":
			// Navigate to selected action
			action := m.actions[m.cursor]
			switch action {
			case "Compile":
				// Create compile view with dependencies and navigate to it
				compileView := compile.NewCompileModel(m.compileService, m.strategy)
				return m, func() tea.Msg {
					return router.NavigateMsg{
						Route: router.RouteStrategyCompile,
						Data:  m.strategy,
						View:  compileView,
					}
				}
			case "Backtest":
				return m, m.router.NavigateWithData(router.RouteStrategyBacktest, m.strategy)
			case "Edit":
				return m, m.router.NavigateWithData(router.RouteStrategyEdit, m.strategy)
			case "Delete":
				return m, m.router.NavigateWithData(router.RouteStrategyDelete, m.strategy)
			}
		}
	case router.NavigateMsg:
		if msg.Route == router.RouteStrategyList {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m strategyDetailView) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"

	// Display strategy metadata
	if m.strategy.Description != "" {
		content += ui.StrategyDescStyle.Render("📝 "+m.strategy.Description) + "\n"
	}

	content += ui.StrategyMetaStyle.Render(fmt.Sprintf("🔗 Exchanges: %v", m.strategy.Exchanges)) + "\n"

	if len(m.strategy.Parameters) > 0 {
		content += ui.StrategyMetaStyle.Render(fmt.Sprintf("⚙️  Parameters: %d", len(m.strategy.Parameters))) + "\n"
	}

	content += "\n" + ui.SubtitleStyle.Render("Select action:") + "\n\n"

	for i, action := range m.actions {
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+action) + "\n"
		} else {
			content += "  " + action + "\n"
		}
	}

	content += "\n" + ui.SubtitleStyle.Render("Enter to select, q to back, ctrl+c to quit")

	return ui.BoxStyle.Render(content)
}
