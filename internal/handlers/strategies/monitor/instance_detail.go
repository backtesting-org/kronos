package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab represents a detail view tab
type Tab int

const (
	TabOverview Tab = iota
	TabPositions
	TabOrderbook
	TabTrades
	TabPnL
)

var tabNames = []string{"Overview", "Positions", "Orderbook", "Trades", "PnL"}

// instanceDetailModel shows detailed view of a single instance
type instanceDetailModel struct {
	ui.BaseModel // Embed for common key handling
	querier      monitoring.ViewQuerier
	instanceID   string
	activeTab    Tab
	loading      bool
	err          error
	width        int
	height       int

	// Cached data
	pnl       *monitoring.PnLView
	metrics   *monitoring.StrategyMetrics
	positions interface{} // TODO: proper type
	orderbook interface{} // TODO: proper type
	trades    interface{} // TODO: proper type
}

// NewInstanceDetailModel creates a detail view for an instance
func NewInstanceDetailModel(querier monitoring.ViewQuerier, instanceID string) tea.Model {
	return &instanceDetailModel{
		BaseModel:  ui.BaseModel{IsRoot: false}, // This is opened from instance list
		querier:    querier,
		instanceID: instanceID,
		activeTab:  TabOverview,
		loading:    true,
	}
}

type detailDataLoadedMsg struct {
	pnl     *monitoring.PnLView
	metrics *monitoring.StrategyMetrics
	err     error
}

type detailTickMsg time.Time

func (m *instanceDetailModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadData(),
		m.tickCmd(),
	)
}

func (m *instanceDetailModel) tickCmd() tea.Cmd {
	interval := 3 * time.Second
	if m.activeTab == TabOrderbook {
		interval = 500 * time.Millisecond
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return detailTickMsg(t)
	})
}

func (m *instanceDetailModel) loadData() tea.Cmd {
	return func() tea.Msg {
		pnl, err := m.querier.QueryPnL(m.instanceID)
		if err != nil {
			return detailDataLoadedMsg{err: err}
		}

		metrics, _ := m.querier.QueryMetrics(m.instanceID)

		return detailDataLoadedMsg{
			pnl:     pnl,
			metrics: metrics,
		}
	}
}

func (m *instanceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case detailDataLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.pnl = msg.pnl
			m.metrics = msg.metrics
		}
		return m, nil

	case detailTickMsg:
		return m, tea.Batch(
			m.loadData(),
			m.tickCmd(),
		)

	case tea.KeyMsg:
		// Handle common keys first (ctrl+c, q, esc)
		if handled, cmd := m.BaseModel.HandleCommonKeys(msg); handled {
			return m, cmd
		}

		switch msg.String() {

		case "left", "h":
			if m.activeTab > 0 {
				m.activeTab--
			}
			return m, nil

		case "right", "l":
			if m.activeTab < TabPnL {
				m.activeTab++
			}
			return m, nil

		case "1":
			m.activeTab = TabOverview
			return m, nil
		case "2":
			m.activeTab = TabPositions
			return m, nil
		case "3":
			m.activeTab = TabOrderbook
			return m, nil
		case "4":
			m.activeTab = TabTrades
			return m, nil
		case "5":
			m.activeTab = TabPnL
			return m, nil

		case "r":
			m.loading = true
			return m, m.loadData()
		}
	}

	return m, nil
}

func (m *instanceDetailModel) View() string {
	var b strings.Builder

	// Header with instance info
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content based on active tab
	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading..."))
	} else if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	} else {
		switch m.activeTab {
		case TabOverview:
			b.WriteString(m.renderOverview())
		case TabPositions:
			b.WriteString(m.renderPositions())
		case TabOrderbook:
			b.WriteString(m.renderOrderbook())
		case TabTrades:
			b.WriteString(m.renderTrades())
		case TabPnL:
			b.WriteString(m.renderPnL())
		}
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[←→] Switch Tab • [1-5] Jump to Tab • [R] Refresh • [Q] Back"))

	return b.String()
}

func (m *instanceDetailModel) renderHeader() string {
	status := "unknown"
	if m.metrics != nil {
		status = m.metrics.Status
	}

	icon := GetStatusIcon(status)
	statusStyle := GetStatusStyle(status)

	title := ui.TitleStyle.Render(strings.ToUpper(m.instanceID))
	statusText := statusStyle.Render(strings.ToUpper(status))

	return fmt.Sprintf("%s  %s %s", title, icon, statusText)
}

func (m *instanceDetailModel) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		if Tab(i) == m.activeTab {
			tabs = append(tabs, TabActiveStyle.Render(fmt.Sprintf("[%s]", name)))
		} else {
			tabs = append(tabs, TabStyle.Render(name))
		}
	}
	return strings.Join(tabs, "  ")
}

func (m *instanceDetailModel) renderOverview() string {
	var b strings.Builder

	// PnL Summary Panel
	pnlContent := m.renderPnLSummary()
	pnlPanel := ui.BoxStyle.Width(35).Render(pnlContent)

	// Quick Stats Panel
	statsContent := m.renderQuickStats()
	statsPanel := ui.BoxStyle.Width(35).Render(statsContent)

	// Side by side
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, pnlPanel, "  ", statsPanel))

	return b.String()
}

func (m *instanceDetailModel) renderPnLSummary() string {
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
	b.WriteString(fmt.Sprintf("Fees:        %s", PnLLossStyle.Render(fmt.Sprintf("-$%.2f", fees))))

	return b.String()
}

func (m *instanceDetailModel) renderQuickStats() string {
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

func (m *instanceDetailModel) renderPositions() string {
	// TODO: Implement positions view
	return ui.SubtitleStyle.Render("Positions view coming soon...")
}

func (m *instanceDetailModel) renderOrderbook() string {
	// TODO: Implement orderbook view
	return ui.SubtitleStyle.Render("Orderbook view coming soon...")
}

func (m *instanceDetailModel) renderTrades() string {
	// TODO: Implement trades view
	return ui.SubtitleStyle.Render("Trades view coming soon...")
}

func (m *instanceDetailModel) renderPnL() string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("PROFIT & LOSS BREAKDOWN"))
	b.WriteString("\n\n")

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
	b.WriteString(m.renderBar(realized, total, maxWidth, ui.ColorSuccess))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%-15s %s\n", "UNREALIZED PNL", FormatPnL(unrealized)))
	b.WriteString(m.renderBar(unrealized, total, maxWidth, ui.ColorWarning))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%-15s %s\n", "TOTAL PNL", FormatPnL(total)))
	b.WriteString(m.renderBar(total, total, maxWidth, ui.ColorPrimary))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Trading Fees:  %s\n", PnLLossStyle.Render(fmt.Sprintf("-$%.2f", fees))))

	return b.String()
}

func (m *instanceDetailModel) renderBar(value, max float64, width int, color lipgloss.Color) string {
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
