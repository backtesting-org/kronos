package live

import (
	"context"
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/services/live"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	strategy *strategy.Strategy
	service  live.LiveService
	starting bool
	started  bool
	err      error
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewLiveModel creates a live trading view
func NewLiveModel(strat *strategy.Strategy, service live.LiveService) tea.Model {
	ctx, cancel := context.WithCancel(context.Background())
	return &liveModel{
		strategy: strat,
		service:  service,
		starting: true,
		started:  false,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (m *liveModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Spawn the live trading instance in background
		// This will start a separate process and return immediately
		err := m.service.ExecuteStrategy(m.ctx, m.strategy, nil)
		return liveSpawnedMsg{err: err}
	}
}

func (m *liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case liveSpawnedMsg:
		m.starting = false
		m.err = msg.err
		if msg.err == nil {
			m.started = true
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			// User can exit immediately - instance runs in background
			return m, bubblon.Cmd(bubblon.Close())
		case "ctrl+c":
			m.cancel()
			return m, bubblon.Cmd(bubblon.Close())
		}
	}
	return m, nil
}

func (m *liveModel) View() string {
	title := ui.TitleStyle.Render("🚀 Live Trading")
	strategyName := ui.StrategyNameStyle.Render(m.strategy.Name)

	var statusSection string
	var helpText string

	if m.starting {
		// Still spawning the process
		statusSection = ui.SubtitleStyle.Render("⏳ Starting live trading instance...")
		helpText = ui.SubtitleStyle.Render("Please wait...")
	} else if m.err != nil {
		// Failed to spawn
		statusIcon := ui.StatusErrorStyle.Render("❌ FAILED TO START")
		errorMsg := ui.StatusErrorStyle.Render(fmt.Sprintf("\n%v", m.err))

		statusSection = lipgloss.JoinVertical(
			lipgloss.Left,
			statusIcon,
			errorMsg,
		)
		helpText = ui.HelpStyle.Render("Press Enter or q to return")
	} else {
		// Successfully spawned - running in background
		statusIcon := ui.StatusReadyStyle.Render("✅ INSTANCE STARTED")
		message := ui.SubtitleStyle.Render("Strategy is now running in the background")

		details := lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Render(
				"• Trading instance spawned as separate process\n" +
					fmt.Sprintf("• Logs: .kronos/instances/%s/stdout.log\n", m.strategy.Name) +
					"• Use 'Monitor' view to check status and metrics\n" +
					"• Instance will continue running after CLI exits",
			)

		statusSection = lipgloss.JoinVertical(
			lipgloss.Left,
			statusIcon,
			"",
			message,
			"",
			details,
		)
		helpText = ui.HelpStyle.Render("Press Enter or q to return to menu")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strategyName,
		"",
		statusSection,
		"",
		helpText,
	)

	return ui.BoxStyle.Render(content)
}

type liveSpawnedMsg struct {
	err error
}
