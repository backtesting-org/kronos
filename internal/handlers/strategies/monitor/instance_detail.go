package monitor

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
	positions *strategy.StrategyExecution
	orderbook *connector.OrderBook
	trades    []connector.Trade

	// Orderbook settings
	selectedAsset string
	depthLevels   int // 5, 10, or 20
}

// NewInstanceDetailModel creates a detail view for an instance
func NewInstanceDetailModel(querier monitoring.ViewQuerier, instanceID string) tea.Model {
	return &instanceDetailModel{
		BaseModel:     ui.BaseModel{IsRoot: false},
		querier:       querier,
		instanceID:    instanceID,
		activeTab:     TabOverview,
		loading:       true,
		selectedAsset: "BTC/USDT", // Default asset
		depthLevels:   10,
	}
}

type detailDataLoadedMsg struct {
	pnl       *monitoring.PnLView
	metrics   *monitoring.StrategyMetrics
	positions *strategy.StrategyExecution
	orderbook *connector.OrderBook
	trades    []connector.Trade
	err       error
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
		var result detailDataLoadedMsg

		// Load PnL
		pnl, err := m.querier.QueryPnL(m.instanceID)
		if err != nil {
			result.err = err
			return result
		}
		result.pnl = pnl

		// Load metrics
		metrics, _ := m.querier.QueryMetrics(m.instanceID)
		result.metrics = metrics

		// Load positions
		positions, _ := m.querier.QueryPositions(m.instanceID)
		result.positions = positions

		// Load orderbook for selected asset
		orderbook, _ := m.querier.QueryOrderbook(m.instanceID, m.selectedAsset)
		result.orderbook = orderbook

		// Load recent trades
		trades, _ := m.querier.QueryRecentTrades(m.instanceID, 20)
		result.trades = trades

		return result
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
			m.positions = msg.positions
			m.orderbook = msg.orderbook
			m.trades = msg.trades
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

		case "d":
			// Toggle depth levels on orderbook tab
			if m.activeTab == TabOrderbook {
				switch m.depthLevels {
				case 5:
					m.depthLevels = 10
				case 10:
					m.depthLevels = 20
				default:
					m.depthLevels = 5
				}
			}
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
	helpText := "[←→] Switch Tab • [1-5] Jump to Tab • [R] Refresh • [Q] Back"
	if m.activeTab == TabOrderbook {
		helpText = "[←→] Switch Tab • [D] Toggle Depth • [R] Refresh • [Q] Back"
	}
	b.WriteString(ui.HelpStyle.Render(helpText))

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
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("POSITIONS"))
	b.WriteString("\n\n")

	if m.positions == nil || (len(m.positions.Orders) == 0 && len(m.positions.Trades) == 0) {
		b.WriteString(ui.SubtitleStyle.Render("No active positions"))
		return b.String()
	}

	// Show orders
	if len(m.positions.Orders) > 0 {
		b.WriteString(TableHeaderStyle.Render(fmt.Sprintf("  %-12s %-8s %-10s %-12s %-12s", "SYMBOL", "SIDE", "QTY", "PRICE", "STATUS")))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 60)))
		b.WriteString("\n")

		for _, order := range m.positions.Orders {
			qty, _ := order.Quantity.Float64()
			price, _ := order.Price.Float64()
			sideStyle := PnLProfitStyle
			if order.Side == connector.OrderSideSell {
				sideStyle = PnLLossStyle
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

func (m *instanceDetailModel) renderOrderbook() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.StrategyNameStyle.Render(fmt.Sprintf("ORDERBOOK - %s", m.selectedAsset)))
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  (Depth: %d levels)", m.depthLevels)))
	b.WriteString("\n\n")

	if m.orderbook == nil {
		b.WriteString(ui.SubtitleStyle.Render("No orderbook data available"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[D] Toggle Depth"))
		return b.String()
	}

	// Calculate max quantity for bar scaling
	maxQty := 0.0
	for i := 0; i < m.depthLevels && i < len(m.orderbook.Asks); i++ {
		qty, _ := m.orderbook.Asks[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	for i := 0; i < m.depthLevels && i < len(m.orderbook.Bids); i++ {
		qty, _ := m.orderbook.Bids[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	if maxQty == 0 {
		maxQty = 1
	}

	barWidth := 30
	askStyle := lipgloss.NewStyle().Foreground(ui.ColorDanger)
	bidStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess)

	// Asks (reversed - lowest ask at bottom)
	b.WriteString(askStyle.Render("                              ASKS"))
	b.WriteString("\n")

	asksToShow := m.depthLevels
	if asksToShow > len(m.orderbook.Asks) {
		asksToShow = len(m.orderbook.Asks)
	}

	// Show asks in reverse order (highest first)
	for i := asksToShow - 1; i >= 0; i-- {
		level := m.orderbook.Asks[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		barLen := int((qty / maxQty) * float64(barWidth))
		if barLen < 1 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)
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
	bidsToShow := m.depthLevels
	if bidsToShow > len(m.orderbook.Bids) {
		bidsToShow = len(m.orderbook.Bids)
	}

	for i := 0; i < bidsToShow; i++ {
		level := m.orderbook.Bids[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		barLen := int((qty / maxQty) * float64(barWidth))
		if barLen < 1 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)
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

	return b.String()
}

func (m *instanceDetailModel) renderTrades() string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("RECENT TRADES"))
	b.WriteString("\n\n")

	if len(m.trades) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No trades yet"))
		return b.String()
	}

	// Table header
	b.WriteString(TableHeaderStyle.Render(fmt.Sprintf("  %-10s %-12s %-8s %-10s %-12s %-10s",
		"TIME", "SYMBOL", "SIDE", "QTY", "PRICE", "FEE")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 70)))
	b.WriteString("\n")

	for _, trade := range m.trades {
		timeStr := trade.Timestamp.Format("15:04:05")
		qty, _ := trade.Quantity.Float64()
		price, _ := trade.Price.Float64()
		fee, _ := trade.Fee.Float64()

		sideStyle := PnLProfitStyle
		if trade.Side == connector.OrderSideSell {
			sideStyle = PnLLossStyle
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
