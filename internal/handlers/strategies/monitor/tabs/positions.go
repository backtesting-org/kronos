package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PositionsModel is a tab that displays positions data
type PositionsModel struct {
	querier    monitoring.ViewQuerier
	instanceID string
	positions  *strategy.StrategyExecution
	loading    bool
	err        error
}

// NewPositionsModel creates a new positions tab
func NewPositionsModel(querier monitoring.ViewQuerier, instanceID string) *PositionsModel {
	return &PositionsModel{
		querier:    querier,
		instanceID: instanceID,
		loading:    true,
	}
}

// Positions messages
type positionsDataMsg struct {
	positions *strategy.StrategyExecution
	err       error
}

type positionsTickMsg time.Time

func (m *PositionsModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.tick(),
	)
}

func (m *PositionsModel) tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return positionsTickMsg(t)
	})
}

func (m *PositionsModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		positions, err := m.querier.QueryPositions(m.instanceID)
		return positionsDataMsg{positions: positions, err: err}
	}
}

func (m *PositionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case positionsDataMsg:
		m.loading = false
		m.err = msg.err
		m.positions = msg.positions
		return m, nil

	case positionsTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.fetchData()
		}
	}
	return m, nil
}

func (m *PositionsModel) View() string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("POSITIONS"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading positions..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if m.positions == nil || (len(m.positions.Orders) == 0 && len(m.positions.Trades) == 0) {
		b.WriteString(ui.SubtitleStyle.Render("No active positions"))
		return b.String()
	}

	// Show orders
	if len(m.positions.Orders) > 0 {
		b.WriteString(tableHeaderStyle.Render(fmt.Sprintf("  %-12s %-8s %-10s %-12s %-12s", "SYMBOL", "SIDE", "QTY", "PRICE", "STATUS")))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 60)))
		b.WriteString("\n")

		for _, order := range m.positions.Orders {
			qty, _ := order.Quantity.Float64()
			price, _ := order.Price.Float64()
			sideStyle := profitStyle
			if order.Side == connector.OrderSideSell {
				sideStyle = lossStyle
			}
			row := fmt.Sprintf("  %-12s %s %-10.4f %-12.2f %-12s",
				order.Symbol,
				sideStyle.Render(fmt.Sprintf("%-8s", order.Side)),
				qty,
				price,
				order.Status,
			)
			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	return b.String()
}
