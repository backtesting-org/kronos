package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TradesModel is a tab that displays recent trades
type TradesModel struct {
	querier    monitoring.ViewQuerier
	instanceID string
	trades     []connector.Trade
	limit      int
	loading    bool
	err        error
}

// NewTradesModel creates a new trades tab
func NewTradesModel(querier monitoring.ViewQuerier, instanceID string) *TradesModel {
	return &TradesModel{
		querier:    querier,
		instanceID: instanceID,
		limit:      20,
		loading:    true,
	}
}

// Trades messages
type tradesDataMsg struct {
	trades []connector.Trade
	err    error
}

type tradesTickMsg time.Time

func (m *TradesModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.tick(),
	)
}

func (m *TradesModel) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tradesTickMsg(t)
	})
}

func (m *TradesModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		trades, err := m.querier.QueryRecentTrades(m.instanceID, m.limit)
		return tradesDataMsg{trades: trades, err: err}
	}
}

func (m *TradesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tradesDataMsg:
		m.loading = false
		m.err = msg.err
		m.trades = msg.trades
		return m, nil

	case tradesTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.fetchData()
		}
	}
	return m, nil
}

func (m *TradesModel) View() string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("RECENT TRADES"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading trades..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if len(m.trades) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No trades yet"))
		return b.String()
	}

	// Table header
	b.WriteString(tableHeaderStyle.Render(fmt.Sprintf("  %-10s %-12s %-8s %-10s %-12s %-10s",
		"TIME", "SYMBOL", "SIDE", "QTY", "PRICE", "FEE")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 70)))
	b.WriteString("\n")

	for _, trade := range m.trades {
		timeStr := trade.Timestamp.Format("15:04:05")
		qty, _ := trade.Quantity.Float64()
		price, _ := trade.Price.Float64()
		fee, _ := trade.Fee.Float64()

		sideStyle := profitStyle
		if trade.Side == connector.OrderSideSell {
			sideStyle = lossStyle
		}

		row := fmt.Sprintf("  %-10s %-12s %s %-10.4f %-12.2f %-10.4f",
			timeStr,
			trade.Symbol,
			sideStyle.Render(fmt.Sprintf("%-8s", trade.Side)),
			qty,
			price,
			fee,
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Showing %d trades", len(m.trades))))

	return b.String()
}
