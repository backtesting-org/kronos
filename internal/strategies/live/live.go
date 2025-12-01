package live

import (
	"context"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/services/live"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

type LiveViewFactory func(*strategy.Strategy) tea.Model

// NewLiveViewFactory creates the factory function for live trading views
func NewLiveViewFactory(
	liveService live.LiveService,
) LiveViewFactory {
	return func(s *strategy.Strategy) tea.Model {
		return NewLiveModel(s, liveService)
	}
}

type liveModel struct {
	strategy  *strategy.Strategy
	service   live.LiveService
	isRunning bool
	err       error
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLiveModel creates a live trading view
func NewLiveModel(strat *strategy.Strategy, service live.LiveService) tea.Model {
	ctx, cancel := context.WithCancel(context.Background())
	return &liveModel{
		strategy:  strat,
		service:   service,
		isRunning: false,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (m *liveModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Start the live trading session
		err := m.service.ExecuteStrategy(m.ctx, m.strategy, nil)
		return liveFinishedMsg{err: err}
	}
}

func (m *liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case liveFinishedMsg:
		m.err = msg.err
		m.isRunning = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancel()
			return m, bubblon.Cmd(bubblon.Close())
		}
	}
	return m, nil
}

func (m *liveModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"
	content += ui.SubtitleStyle.Render("Live Trading") + "\n\n"

	if m.err != nil {
		content += ui.StatusErrorStyle.Render("❌ Trading Error") + "\n\n"
		content += ui.StatusErrorStyle.Render(m.err.Error()) + "\n"
		content += "\n" + ui.SubtitleStyle.Render("Press Ctrl+C to go back")
	} else {
		content += "🚀 Trading live on exchange...\n\n"
		content += "Press Ctrl+C to stop\n"
	}

	return ui.BoxStyle.Render(content)
}

type liveFinishedMsg struct {
	err error
}
