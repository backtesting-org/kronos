package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Orderbook view styles
var (
	askStyle = lipgloss.NewStyle().Foreground(ui.ColorDanger)
	bidStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess)
)

// OrderbookModel is a tab that displays live orderbook data
type OrderbookModel struct {
	querier    monitoring.ViewQuerier
	instanceID string
	depth      int
	orderbook  *connector.OrderBook
	loading    bool
	err        error

	// Available asset/exchange pairs
	availableAssets []monitoring.AssetExchange
	selectedIndex   int
}

// NewOrderbookModel creates a new orderbook tab
func NewOrderbookModel(querier monitoring.ViewQuerier, instanceID string) *OrderbookModel {
	return &OrderbookModel{
		querier:         querier,
		instanceID:      instanceID,
		depth:           10,
		loading:         true,
		availableAssets: []monitoring.AssetExchange{},
		selectedIndex:   0,
	}
}

// Orderbook messages
type orderbookDataMsg struct {
	orderbook *connector.OrderBook
	err       error
}

type orderbookAssetsMsg struct {
	assets []monitoring.AssetExchange
	err    error
}

type orderbookTickMsg time.Time

func (m *OrderbookModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchAssets(),
		m.tick(),
	)
}

func (m *OrderbookModel) tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return orderbookTickMsg(t)
	})
}

func (m *OrderbookModel) fetchAssets() tea.Cmd {
	return func() tea.Msg {
		assets, err := m.querier.QueryAvailableAssets(m.instanceID)
		return orderbookAssetsMsg{assets: assets, err: err}
	}
}

func (m *OrderbookModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		if len(m.availableAssets) == 0 {
			return orderbookDataMsg{err: fmt.Errorf("no assets available")}
		}
		selected := m.availableAssets[m.selectedIndex]
		orderbook, err := m.querier.QueryOrderbook(m.instanceID, selected.Asset, selected.Exchange)
		return orderbookDataMsg{orderbook: orderbook, err: err}
	}
}

func (m *OrderbookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case orderbookAssetsMsg:
		if msg.err == nil && len(msg.assets) > 0 {
			m.availableAssets = msg.assets
			m.selectedIndex = 0
			// Now fetch orderbook for first asset
			return m, m.fetchData()
		}
		m.loading = false
		m.err = msg.err
		return m, nil

	case orderbookDataMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.orderbook = msg.orderbook
		}
		return m, nil

	case orderbookTickMsg:
		if len(m.availableAssets) > 0 {
			return m, tea.Batch(m.fetchData(), m.tick())
		}
		return m, m.tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			// Toggle depth
			switch m.depth {
			case 5:
				m.depth = 10
			case 10:
				m.depth = 20
			default:
				m.depth = 5
			}
			return m, nil

		case "tab", "n":
			// Next asset/exchange
			if len(m.availableAssets) > 0 {
				m.selectedIndex = (m.selectedIndex + 1) % len(m.availableAssets)
				m.loading = true
				m.orderbook = nil
				return m, m.fetchData()
			}
			return m, nil

		case "shift+tab", "p":
			// Previous asset/exchange
			if len(m.availableAssets) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.availableAssets) - 1
				}
				m.loading = true
				m.orderbook = nil
				return m, m.fetchData()
			}
			return m, nil

		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.fetchAssets(), m.fetchData())
		}
	}
	return m, nil
}

func (m *OrderbookModel) View() string {
	var b strings.Builder

	// Header with asset selector
	if len(m.availableAssets) > 0 {
		selected := m.availableAssets[m.selectedIndex]
		b.WriteString(ui.StrategyNameStyle.Render(fmt.Sprintf("ORDERBOOK - %s @ %s", selected.Asset, selected.Exchange)))
		b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  (%d/%d)", m.selectedIndex+1, len(m.availableAssets))))
	} else {
		b.WriteString(ui.StrategyNameStyle.Render("ORDERBOOK"))
	}
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  Depth: %d", m.depth)))
	b.WriteString("\n\n")

	if m.loading && m.orderbook == nil && m.err == nil {
		b.WriteString(ui.SubtitleStyle.Render("Loading orderbook..."))
		return b.String()
	}

	if len(m.availableAssets) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No trading assets configured"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[R] Refresh"))
		return b.String()
	}

	if m.orderbook == nil {
		if m.err != nil {
			b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("No data: %v", m.err)))
		} else {
			b.WriteString(ui.SubtitleStyle.Render("No orderbook data available"))
		}
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[Tab] Next Asset • [D] Depth • [R] Retry"))
		return b.String()
	}

	// Calculate max quantity for bar scaling
	maxQty := m.calculateMaxQuantity()
	if maxQty == 0 {
		maxQty = 1
	}

	barWidth := 30

	// Asks (reversed - lowest ask at bottom)
	b.WriteString(askStyle.Render("                              ASKS"))
	b.WriteString("\n")

	asksToShow := m.depth
	if asksToShow > len(m.orderbook.Asks) {
		asksToShow = len(m.orderbook.Asks)
	}

	// Show asks in reverse order (highest first)
	for i := asksToShow - 1; i >= 0; i-- {
		level := m.orderbook.Asks[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		bar := renderDepthBar(qty, maxQty, barWidth)
		row := fmt.Sprintf("  %12.2f  %s  %8.4f", price, askStyle.Render(bar), qty)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Spread
	if len(m.orderbook.Asks) > 0 && len(m.orderbook.Bids) > 0 {
		bestAsk, _ := m.orderbook.Asks[0].Price.Float64()
		bestBid, _ := m.orderbook.Bids[0].Price.Float64()
		spread := bestAsk - bestBid
		spreadPct := (spread / bestBid) * 100

		spreadLine := fmt.Sprintf("  ──────────── SPREAD: $%.2f (%.3f%%) ────────────", spread, spreadPct)
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(spreadLine))
		b.WriteString("\n")
	}

	// Bids
	bidsToShow := m.depth
	if bidsToShow > len(m.orderbook.Bids) {
		bidsToShow = len(m.orderbook.Bids)
	}

	for i := 0; i < bidsToShow; i++ {
		level := m.orderbook.Bids[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		bar := renderDepthBar(qty, maxQty, barWidth)
		row := fmt.Sprintf("  %12.2f  %s  %8.4f", price, bidStyle.Render(bar), qty)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString(bidStyle.Render("                              BIDS"))
	b.WriteString("\n\n")

	// Mid price info
	if len(m.orderbook.Asks) > 0 && len(m.orderbook.Bids) > 0 {
		bestAsk, _ := m.orderbook.Asks[0].Price.Float64()
		bestBid, _ := m.orderbook.Bids[0].Price.Float64()
		midPrice := (bestAsk + bestBid) / 2

		b.WriteString(fmt.Sprintf("Mid: $%.2f   Bid: $%.2f   Ask: $%.2f",
			midPrice, bestBid, bestAsk))
	}

	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[Tab] Next Asset • [D] Toggle Depth"))

	return b.String()
}

func (m *OrderbookModel) calculateMaxQuantity() float64 {
	maxQty := 0.0
	for i := 0; i < m.depth && i < len(m.orderbook.Asks); i++ {
		qty, _ := m.orderbook.Asks[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	for i := 0; i < m.depth && i < len(m.orderbook.Bids); i++ {
		qty, _ := m.orderbook.Bids[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	return maxQty
}

func renderDepthBar(qty, maxQty float64, width int) string {
	barLen := int((qty / maxQty) * float64(width))
	if barLen < 1 {
		barLen = 1
	}
	return strings.Repeat("█", barLen) + strings.Repeat("░", width-barLen)
}
