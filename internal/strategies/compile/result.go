package compile

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type resultModel struct {
	strategy *strategy.Strategy
	err      error
}

// NewResultModel creates a result model that shows compilation result
func NewResultModel(strat *strategy.Strategy, err error) tea.Model {
	return &resultModel{
		strategy: strat,
		err:      err,
	}
}

func (m *resultModel) Init() tea.Cmd {
	return nil
}

func (m *resultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			// Return to detail view (parent) by closing this result view
			return m, bubblon.Cmd(bubblon.Close())
		}
	}
	return m, nil
}

func (m *resultModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"

	if m.err == nil {
		content += ui.StatusReadyStyle.Render("✅ Compilation Successful") + "\n\n"
		content += "Strategy has been compiled to .so plugin\n"
		content += "Ready for backtest or live trading\n"
	} else {
		content += ui.StatusErrorStyle.Render("❌ Compilation Failed") + "\n\n"
		content += ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	content += "\n" + ui.SubtitleStyle.Render("Press Enter or q to go back")

	return ui.BoxStyle.Render(content)
}
