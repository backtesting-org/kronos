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

// PnLModel is a tab that displays PnL data
type PnLModel struct {
	querier    monitoring.ViewQuerier
	instanceID string
	pnl        *monitoring.PnLView
	loading    bool
	err        error
}

// NewPnLModel creates a new PnL tab
func NewPnLModel(querier monitoring.ViewQuerier, instanceID string) *PnLModel {
	return &PnLModel{
		querier:    querier,
		instanceID: instanceID,
		loading:    true,
	}
}

// PnL messages
type pnlDataMsg struct {
	pnl *monitoring.PnLView
	err error
}

type pnlTickMsg time.Time

func (m *PnLModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.tick(),
	)
}

func (m *PnLModel) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return pnlTickMsg(t)
	})
}

func (m *PnLModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		pnl, err := m.querier.QueryPnL(m.instanceID)
		return pnlDataMsg{pnl: pnl, err: err}
	}
}

func (m *PnLModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pnlDataMsg:
		m.loading = false
		m.err = msg.err
		m.pnl = msg.pnl
		return m, nil

	case pnlTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.fetchData()
		}
	}
	return m, nil
}

func (m *PnLModel) View() string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("PROFIT & LOSS BREAKDOWN"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading PnL..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if m.pnl == nil {
		b.WriteString(ui.SubtitleStyle.Render("No PnL data available"))
		return b.String()
	}

	realized, _ := m.pnl.RealizedPnL.Float64()
	unrealized, _ := m.pnl.UnrealizedPnL.Float64()
	total, _ := m.pnl.TotalPnL.Float64()
	fees, _ := m.pnl.TotalFees.Float64()

	// Visual bars
	maxWidth := 40

	b.WriteString(fmt.Sprintf("%-15s %s\n", "REALIZED PNL", FormatPnL(realized)))
	b.WriteString(renderBar(realized, total, maxWidth, ui.ColorSuccess))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%-15s %s\n", "UNREALIZED PNL", FormatPnL(unrealized)))
	b.WriteString(renderBar(unrealized, total, maxWidth, ui.ColorWarning))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%-15s %s\n", "TOTAL PNL", FormatPnL(total)))
	b.WriteString(renderBar(total, total, maxWidth, ui.ColorPrimary))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Trading Fees:  %s\n", lossStyle.Render(fmt.Sprintf("-$%.2f", fees))))

	return b.String()
}

func renderBar(value, max float64, width int, color lipgloss.Color) string {
	if max == 0 {
		max = 1
	}
	ratio := value / max
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	filled := int(float64(width) * ratio)
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}
