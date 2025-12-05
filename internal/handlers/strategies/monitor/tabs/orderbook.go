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

	// Live indicator styles
	liveStyle = lipgloss.NewStyle().
		Foreground(ui.ColorSuccess).
		Bold(true)

	pulseStyle = lipgloss.NewStyle().
		Foreground(ui.ColorWarning).
		Bold(true)
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

	// Live update tracking
	lastUpdate  time.Time
	updateCount int
	showPulse   bool
	lastBestBid float64
	lastBestAsk float64
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
		updateCount:     0,
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
type orderbookPulseOffMsg struct{}

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

func (m *OrderbookModel) pulseOff() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return orderbookPulseOffMsg{}
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
			return m, m.fetchData()
		}
		m.loading = false
		m.err = msg.err
		return m, nil

	case orderbookDataMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil && msg.orderbook != nil {
			// Check if data actually changed
			changed := m.hasDataChanged(msg.orderbook)
			m.orderbook = msg.orderbook
			m.lastUpdate = time.Now()
			m.updateCount++

			if changed {
				m.showPulse = true
				// Update last prices
				if len(msg.orderbook.Bids) > 0 {
					m.lastBestBid, _ = msg.orderbook.Bids[0].Price.Float64()
				}
				if len(msg.orderbook.Asks) > 0 {
					m.lastBestAsk, _ = msg.orderbook.Asks[0].Price.Float64()
				}
				return m, m.pulseOff()
			}
		}
		return m, nil

	case orderbookPulseOffMsg:
		m.showPulse = false
		return m, nil

	case orderbookTickMsg:
		if len(m.availableAssets) > 0 {
			return m, tea.Batch(m.fetchData(), m.tick())
		}
		return m, m.tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
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
			if len(m.availableAssets) > 0 {
				m.selectedIndex = (m.selectedIndex + 1) % len(m.availableAssets)
				m.loading = true
				m.orderbook = nil
				m.updateCount = 0
				return m, m.fetchData()
			}
			return m, nil

		case "shift+tab", "p":
			if len(m.availableAssets) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.availableAssets) - 1
				}
				m.loading = true
				m.orderbook = nil
				m.updateCount = 0
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

// hasDataChanged checks if the orderbook data has meaningfully changed
func (m *OrderbookModel) hasDataChanged(newOB *connector.OrderBook) bool {
	if m.orderbook == nil {
		return true
	}
	if len(newOB.Bids) > 0 && len(m.orderbook.Bids) > 0 {
		newBid, _ := newOB.Bids[0].Price.Float64()
		if newBid != m.lastBestBid {
			return true
		}
	}
	if len(newOB.Asks) > 0 && len(m.orderbook.Asks) > 0 {
		newAsk, _ := newOB.Asks[0].Price.Float64()
		if newAsk != m.lastBestAsk {
			return true
		}
	}
	return false
}

func (m *OrderbookModel) View() string {
	var b strings.Builder

	// Header with asset selector and live indicator
	b.WriteString(m.renderHeader())
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

	// Render orderbook
	b.WriteString(m.renderOrderbook())

	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[Tab] Next Asset • [D] Toggle Depth"))

	return b.String()
}

func (m *OrderbookModel) renderHeader() string {
	var header strings.Builder

	// Asset/exchange info
	if len(m.availableAssets) > 0 {
		selected := m.availableAssets[m.selectedIndex]
		header.WriteString(ui.StrategyNameStyle.Render(fmt.Sprintf("ORDERBOOK - %s @ %s", selected.Asset, selected.Exchange)))
		header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  (%d/%d)", m.selectedIndex+1, len(m.availableAssets))))
	} else {
		header.WriteString(ui.StrategyNameStyle.Render("ORDERBOOK"))
	}

	header.WriteString("  ")

	// Live indicator with pulse
	if m.showPulse {
		header.WriteString(pulseStyle.Render("◉ LIVE"))
	} else if !m.lastUpdate.IsZero() {
		header.WriteString(liveStyle.Render("● LIVE"))
	}

	// Update stats
	if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate)
		if ago < time.Second {
			header.WriteString(ui.SubtitleStyle.Render("  <1s ago"))
		} else {
			header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  %ds ago", int(ago.Seconds()))))
		}
		header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  |  %d updates", m.updateCount)))
	}

	header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  |  Depth: %d", m.depth)))

	return header.String()
}

func (m *OrderbookModel) renderOrderbook() string {
	var b strings.Builder

	maxQty := m.calculateMaxQuantity()
	if maxQty == 0 {
		maxQty = 1
	}

	barWidth := 30

	// Use pulse style for the labels if we just updated
	currentAskStyle := askStyle
	currentBidStyle := bidStyle
	if m.showPulse {
		currentAskStyle = pulseStyle
		currentBidStyle = pulseStyle
	}

	// Asks header
	b.WriteString(currentAskStyle.Render("                              ASKS"))
	b.WriteString("\n")

	asksToShow := m.depth
	if asksToShow > len(m.orderbook.Asks) {
		asksToShow = len(m.orderbook.Asks)
	}

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

	// Bids footer
	b.WriteString(currentBidStyle.Render("                              BIDS"))
	b.WriteString("\n\n")

	// Mid price info
	if len(m.orderbook.Asks) > 0 && len(m.orderbook.Bids) > 0 {
		bestAsk, _ := m.orderbook.Asks[0].Price.Float64()
		bestBid, _ := m.orderbook.Bids[0].Price.Float64()
		midPrice := (bestAsk + bestBid) / 2

		b.WriteString(fmt.Sprintf("Mid: $%.2f   Bid: $%.2f   Ask: $%.2f",
			midPrice, bestBid, bestAsk))
	}

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
