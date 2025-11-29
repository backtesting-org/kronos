package browse

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type StrategyDetailView interface {
	tea.Model
}

// strategyDetailView represents the strategy detail view with action options (STRATEGY screen)
type strategyDetailView struct {
	strategy       *strategy.Strategy
	actions        []string
	cursor         int
	compileService shared.CompileService
}

// newStrategyDetailView is the private constructor called by the factory
func newStrategyDetailView(
	compileService shared.CompileService,
	s *strategy.Strategy,
) tea.Model {
	return &strategyDetailView{
		strategy:       s,
		actions:        []string{"Compile", "Backtest", "Edit", "Delete"},
		cursor:         0,
		compileService: compileService,
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
			case "Compile":
				// TODO: Route to compile view with strategy data
				return m, nil
			case "Backtest":
				// TODO: Route to backtest view with strategy data
				return m, nil
			case "Edit":
				// TODO: Route to edit view with strategy data
				return m, nil
			case "Delete":
				// TODO: Handle delete action
				return m, nil
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
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+action) + "\n"
		} else {
			content += "  " + action + "\n"
		}
	}

	content += "\n" + ui.SubtitleStyle.Render("Enter to select, q to back, ctrl+c to quit")

	return ui.BoxStyle.Render(content)
}
