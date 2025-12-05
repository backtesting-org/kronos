package compile

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

type CompileModel interface {
	tea.Model
	SetStrategy(strategy *strategy.Strategy)
	Done() bool
}

type compileModel struct {
	strategy       *strategy.Strategy
	compileService strategyTypes.CompileService
	done           bool
	err            error
	output         string
	progressValue  float64
	frame          int
}

// NewCompileModel creates a compile view with all dependencies
func NewCompileModel(compileService strategyTypes.CompileService) CompileModel {
	return &compileModel{
		strategy:       nil,
		compileService: compileService,
		done:           false,
		progressValue:  0.0,
		frame:          0,
	}
}

func (m *compileModel) SetStrategy(strategy *strategy.Strategy) {
	m.strategy = strategy
}

// progressTickMsg is sent periodically to animate the progress bar
type progressTickMsg time.Time

func tickProgress() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return progressTickMsg(t)
	})
}

// renderProgressBar creates a simple ASCII progress bar
func renderProgressBar(percent float64, width int) string {
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)

	percentStr := fmt.Sprintf("%.0f%%", percent*100)

	progressStyle := lipgloss.NewStyle().
		Foreground(ui.ColorSuccess).
		Bold(true)

	emptyStyle := lipgloss.NewStyle().
		Foreground(ui.ColorMuted)

	return progressStyle.Render(bar) + emptyStyle.Render(empty) + " " + percentStr
}

func (m *compileModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			// Run compile in background
			err := m.compileService.CompileStrategy(m.strategy.Path)
			return CompileFinishedMsg{Err: err}
		},
		tickProgress(),
	)
}

func (m *compileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CompileFinishedMsg:
		m.done = true
		m.err = msg.Err
		m.progressValue = 1.0
		// When compile finishes, replace with result view
		resultView := NewResultModel(m.strategy, m.err)
		return m, bubblon.Replace(resultView)

	case progressTickMsg:
		if !m.done {
			// Increment progress (simulate activity)
			m.progressValue += 0.02
			if m.progressValue > 0.95 {
				m.progressValue = 0.1 // Reset to simulate ongoing work
			}
			m.frame++
			return m, tickProgress()
		}
		return m, nil

	case tea.KeyMsg:
		// Don't allow interaction during compile
		return m, nil
	}
	return m, nil
}

func (m *compileModel) View() string {
	// Title section
	title := ui.TitleStyle.Render("🔨 Compiling Strategy")
	strategyName := ui.StrategyNameStyle.Render(m.strategy.Name)

	// Status message
	status := ui.SubtitleStyle.Render("Building plugin binary...")

	// Progress bar (width 50 characters)
	progressBar := renderProgressBar(m.progressValue, 50)

	// Animated spinner based on frame
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinners[m.frame%len(spinners)]
	activity := ui.SubtitleStyle.Render(spinner + " Working...")

	// Build content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strategyName,
		"",
		status,
		"",
		progressBar,
		"",
		activity,
	)

	return ui.BoxStyle.Render(content)
}

// Done returns whether compilation is complete
func (m *compileModel) Done() bool {
	return m.done
}

// GetStrategy returns the strategy being compiled
func (m *compileModel) GetStrategy() *strategy.Strategy {
	return m.strategy
}

// GetError returns any compilation error
func (m *compileModel) GetError() error {
	return m.err
}

// CompileFinishedMsg is sent when compilation completes
type CompileFinishedMsg struct {
	Err error
}
