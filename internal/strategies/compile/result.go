package compile

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type ResultModel struct {
	strategy     *strategy.Strategy
	err          error
	backToDetail bool
}

// NewResultModel creates a result model - called directly when compile finishes
func NewResultModel(strat *strategy.Strategy, err error) ResultModel {
	return ResultModel{
		strategy:     strat,
		err:          err,
		backToDetail: false,
	}
}

func (m ResultModel) Init() tea.Cmd {
	return nil
}

func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			m.backToDetail = true
		}
	}
	return m, nil
}

func (m ResultModel) View() string {
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

// ShouldBackToDetail returns whether to navigate back to detail
func (m ResultModel) ShouldBackToDetail() bool {
	return m.backToDetail
}

// Reset clears the back flag
func (m ResultModel) Reset() {
	m.backToDetail = false
}
