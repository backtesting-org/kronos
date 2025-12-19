package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfilingModel displays strategy execution profiling data
type ProfilingModel struct {
	querier    monitoring.ViewQuerier
	instanceID string

	// Data
	stats      *monitoring.ProfilingStats
	executions []monitoring.ProfilingMetrics

	// UI state
	loading  bool
	err      error
	cursor   int // For scrolling through executions
	viewMode int // 0 = stats, 1 = recent executions
}

// NewProfilingModel creates a new profiling tab
func NewProfilingModel(querier monitoring.ViewQuerier, instanceID string) *ProfilingModel {
	return &ProfilingModel{
		querier:    querier,
		instanceID: instanceID,
		loading:    true,
		viewMode:   0,
	}
}

// Messages
type profilingDataMsg struct {
	stats      *monitoring.ProfilingStats
	executions []monitoring.ProfilingMetrics
	err        error
}

type profilingTickMsg time.Time

func (m *ProfilingModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.tick(),
	)
}

func (m *ProfilingModel) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return profilingTickMsg(t)
	})
}

func (m *ProfilingModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.querier.QueryProfilingStats(m.instanceID)
		if err != nil {
			return profilingDataMsg{err: err}
		}

		executions, err := m.querier.QueryRecentExecutions(m.instanceID, 50)
		if err != nil {
			return profilingDataMsg{stats: stats, err: err}
		}

		return profilingDataMsg{
			stats:      stats,
			executions: executions,
		}
	}
}

func (m *ProfilingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case profilingDataMsg:
		m.loading = false
		m.err = msg.err
		m.stats = msg.stats
		m.executions = msg.executions
		return m, nil

	case profilingTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, m.fetchData()
		case "tab":
			m.viewMode = (m.viewMode + 1) % 2
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.executions)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *ProfilingModel) View() string {
	if m.loading {
		return ui.SubtitleStyle.Render("Loading profiling data...")
	}

	if m.err != nil {
		return ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.viewMode == 0 {
		return m.renderStats()
	}
	return m.renderExecutions()
}

func (m *ProfilingModel) renderStats() string {
	if m.stats == nil {
		return "No profiling data available"
	}

	var b strings.Builder

	// Header
	b.WriteString(ui.TitleStyle.Render("📊 Profiling Statistics"))
	b.WriteString("\n\n")

	// Summary Stats
	b.WriteString(ui.SubtitleStyle.Render("Execution Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total Runs:     %d\n", m.stats.TotalRuns))
	b.WriteString(fmt.Sprintf("Success:        %d (%.1f%%)\n", m.stats.SuccessCount, m.stats.SuccessRate))
	b.WriteString(fmt.Sprintf("Failures:       %d\n", m.stats.FailureCount))
	b.WriteString(fmt.Sprintf("Last Execution: %s\n", m.stats.LastExecution.Format("15:04:05")))
	b.WriteString("\n")

	// Timing Statistics
	b.WriteString(ui.SubtitleStyle.Render("Execution Time"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Average:        %v\n", m.stats.AvgDuration.Round(time.Microsecond)))
	b.WriteString(fmt.Sprintf("Min:            %v\n", m.stats.MinDuration.Round(time.Microsecond)))
	b.WriteString(fmt.Sprintf("Max:            %v\n", m.stats.MaxDuration.Round(time.Microsecond)))
	b.WriteString("\n")

	// Percentiles
	b.WriteString(ui.SubtitleStyle.Render("Percentiles"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("P50 (median):   %v\n", m.stats.P50.Round(time.Microsecond)))
	b.WriteString(fmt.Sprintf("P95:            %v\n", m.stats.P95.Round(time.Microsecond)))
	b.WriteString(fmt.Sprintf("P99:            %v\n", m.stats.P99.Round(time.Microsecond)))
	b.WriteString("\n")

	// Performance indicator
	avgMs := m.stats.AvgDuration.Milliseconds()
	var perfIndicator string
	if avgMs < 50 {
		perfIndicator = ui.StatusReadyStyle.Render("🟢 Excellent")
	} else if avgMs < 100 {
		perfIndicator = ui.StatusReadyStyle.Render("🟡 Good")
	} else if avgMs < 500 {
		perfIndicator = "🟠 Fair"
	} else {
		perfIndicator = ui.StatusErrorStyle.Render("🔴 Slow")
	}
	b.WriteString(fmt.Sprintf("Performance: %s\n", perfIndicator))

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("Press [tab] to view recent executions • [r] to refresh"))

	return b.String()
}

func (m *ProfilingModel) renderExecutions() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.TitleStyle.Render("📋 Recent Executions"))
	b.WriteString("\n\n")

	if len(m.executions) == 0 {
		b.WriteString("No execution data available\n")
		return b.String()
	}

	// Table header
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	b.WriteString(headerStyle.Render(fmt.Sprintf("%-19s %-12s %-8s %-15s", "Time", "Duration", "Status", "Indicators")))
	b.WriteString("\n")

	// Show last 10 executions
	start := 0
	if len(m.executions) > 10 {
		start = len(m.executions) - 10
	}

	for i := start; i < len(m.executions); i++ {
		exec := m.executions[i]

		// Cursor indicator
		cursor := "  "
		if i == m.cursor {
			cursor = "→ "
		}

		// Status indicator
		var status string
		if exec.Success {
			status = ui.StatusReadyStyle.Render("✓")
		} else {
			status = ui.StatusErrorStyle.Render("✗")
		}

		// Format time
		timeStr := exec.Timestamp.Format("15:04:05")

		// Duration with color coding
		durationStr := exec.ExecutionTime.Round(time.Millisecond).String()
		durationStyle := lipgloss.NewStyle()
		if exec.ExecutionTime.Milliseconds() > 100 {
			durationStyle = durationStyle.Foreground(lipgloss.Color("208")) // Orange
		}
		if exec.ExecutionTime.Milliseconds() > 500 {
			durationStyle = durationStyle.Foreground(lipgloss.Color("196")) // Red
		}

		// Indicator count
		indicatorCount := len(exec.IndicatorMetrics)
		indicatorStr := fmt.Sprintf("%d indicators", indicatorCount)

		line := fmt.Sprintf("%s%-19s %-12s %s %-15s",
			cursor,
			timeStr,
			durationStyle.Render(durationStr),
			status,
			indicatorStr,
		)

		b.WriteString(line)
		b.WriteString("\n")

		// Show error if failed and selected
		if !exec.Success && i == m.cursor && exec.Error != "" {
			errorLine := fmt.Sprintf("  Error: %s", exec.Error)
			b.WriteString(ui.StatusErrorStyle.Render(errorLine))
			b.WriteString("\n")
		}

		// Show indicator breakdown if selected
		if i == m.cursor && len(exec.IndicatorMetrics) > 0 {
			b.WriteString("  Indicators:\n")
			for name, timing := range exec.IndicatorMetrics {
				b.WriteString(fmt.Sprintf("    • %s: %v (%d calls)\n",
					name,
					timing.Duration.Round(time.Microsecond),
					timing.Calls))
			}
			if exec.SignalGenTime > 0 {
				b.WriteString(fmt.Sprintf("    • Signal Generation: %v\n", exec.SignalGenTime.Round(time.Microsecond)))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("↑/↓ to navigate • [tab] to view stats • [r] to refresh"))

	return b.String()
}
