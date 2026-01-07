package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	monitoring2 "github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverviewModel is a tab that displays an overview of the strategy
type OverviewModel struct {
	querier    monitoring2.ViewQuerier
	instanceID string
	pnl        *monitoring2.PnLView
	metrics    *monitoring2.StrategyMetrics
	loading    bool
	err        error
}

// NewOverviewModel creates a new overview tab
func NewOverviewModel(querier monitoring2.ViewQuerier, instanceID string) *OverviewModel {
	return &OverviewModel{
		querier:    querier,
		instanceID: instanceID,
		loading:    true,
	}
}

// Overview messages
type overviewDataMsg struct {
	pnl     *monitoring2.PnLView
	metrics *monitoring2.StrategyMetrics
	err     error
}

type overviewTickMsg time.Time

func (m *OverviewModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.tick(),
	)
}

func (m *OverviewModel) tick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return overviewTickMsg(t)
	})
}

func (m *OverviewModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		pnl, err := m.querier.QueryPnL(m.instanceID)
		if err != nil {
			return overviewDataMsg{err: err}
		}
		metrics, _ := m.querier.QueryMetrics(m.instanceID)
		return overviewDataMsg{pnl: pnl, metrics: metrics}
	}
}

func (m *OverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case overviewDataMsg:
		m.loading = false
		m.err = msg.err
		m.pnl = msg.pnl
		m.metrics = msg.metrics
		return m, nil

	case overviewTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.fetchData()
		}
	}
	return m, nil
}

func (m *OverviewModel) View() string {
	if m.loading {
		return ui.SubtitleStyle.Render("Loading overview...")
	}

	if m.err != nil {
		return ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// PnL Summary Panel
	pnlContent := m.renderPnLSummary()
	pnlPanel := ui.BoxStyle.Width(35).Render(pnlContent)

	// Quick Stats Panel
	statsContent := m.renderQuickStats()
	statsPanel := ui.BoxStyle.Width(35).Render(statsContent)

	// Side by side
	return lipgloss.JoinHorizontal(lipgloss.Top, pnlPanel, "  ", statsPanel)
}

func (m *OverviewModel) renderPnLSummary() string {
	var b strings.Builder
	b.WriteString(ui.StrategyNameStyle.Render("PNL SUMMARY"))
	b.WriteString("\n\n")

	if m.pnl == nil {
		b.WriteString(ui.SubtitleStyle.Render("No data"))
		return b.String()
	}

	realized, _ := m.pnl.RealizedPnL.Float64()
	unrealized, _ := m.pnl.UnrealizedPnL.Float64()
	total, _ := m.pnl.TotalPnL.Float64()
	fees, _ := m.pnl.TotalFees.Float64()

	b.WriteString(fmt.Sprintf("Realized:    %s\n", FormatPnL(realized)))
	b.WriteString(fmt.Sprintf("Unrealized:  %s\n", FormatPnL(unrealized)))
	b.WriteString(fmt.Sprintf("Total:       %s\n", FormatPnL(total)))
	b.WriteString(fmt.Sprintf("Fees:        %s", lossStyle.Render(fmt.Sprintf("-$%.2f", fees))))

	return b.String()
}

func (m *OverviewModel) renderQuickStats() string {
	var b strings.Builder
	b.WriteString(ui.StrategyNameStyle.Render("QUICK STATS"))
	b.WriteString("\n\n")

	if m.metrics == nil {
		b.WriteString(ui.SubtitleStyle.Render("No data"))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Signals Generated:  %d\n", m.metrics.SignalsGenerated))
	b.WriteString(fmt.Sprintf("Signals Executed:   %d\n", m.metrics.SignalsExecuted))

	successRate := 0.0
	if m.metrics.SignalsGenerated > 0 {
		successRate = float64(m.metrics.SignalsExecuted) / float64(m.metrics.SignalsGenerated) * 100
	}
	b.WriteString(fmt.Sprintf("Success Rate:       %.0f%%\n", successRate))
	b.WriteString(fmt.Sprintf("Avg Latency:        %v", m.metrics.AverageLatency))

	return b.String()
}
