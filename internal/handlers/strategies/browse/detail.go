package browse

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/compile"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/live"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type ActionType int

const (
	ActionCompile ActionType = iota
	ActionStartTrading
)

var actionNames = map[ActionType]string{
	ActionCompile:      "Compile",
	ActionStartTrading: "Start Trading",
}

type StrategyDetailView interface {
	tea.Model
}

// strategyDetailView represents the strategy detail view with action options (STRATEGY screen)
type strategyDetailView struct {
	strategy       *strategy.Strategy
	actions        []ActionType
	cursor         int
	compileFactory compile.CompileViewFactory
	liveFactory    live.LiveViewFactory
}

// newStrategyDetailView is the private constructor called by the factory
func newStrategyDetailView(
	compileFactory compile.CompileViewFactory,
	liveFactory live.LiveViewFactory,
	s *strategy.Strategy,
) tea.Model {
	return &strategyDetailView{
		strategy:       s,
		actions:        []ActionType{ActionCompile, ActionStartTrading},
		cursor:         0,
		compileFactory: compileFactory,
		liveFactory:    liveFactory,
	}
}

func (m *strategyDetailView) Init() tea.Cmd {
	return nil
}

func (m *strategyDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Pop back to list view using Bubblon
			return m, bubblon.Cmd(bubblon.Close())
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
			case ActionCompile:
				compileView := m.compileFactory(m.strategy)
				return m, bubblon.Open(compileView)
			case ActionStartTrading:
				liveView := m.liveFactory(m.strategy)
				return m, bubblon.Open(liveView)
			}
		}
	}
	return m, nil
}

func (m *strategyDetailView) View() string {
	if m.strategy == nil {
		return ui.BoxStyle.Render("Strategy not found")
	}

	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"
	content += ui.SubtitleStyle.Render("Select action:") + "\n\n"

	for i, action := range m.actions {
		actionName := actionNames[action]
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+actionName) + "\n"
		} else {
			content += "  " + actionName + "\n"
		}
	}

	content += "\n" + ui.SubtitleStyle.Render("Enter to select, q to back, ctrl+c to quit")

	return ui.BoxStyle.Render(content)
}
