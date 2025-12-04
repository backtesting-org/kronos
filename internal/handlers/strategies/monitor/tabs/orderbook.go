package tabs

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/charmbracelet/lipgloss"
)

// Orderbook view styles
var (
	askStyle = lipgloss.NewStyle().Foreground(ui.ColorDanger)
	bidStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess)
)

// RenderOrderbook renders the orderbook view
func RenderOrderbook(orderbook *connector.OrderBook, selectedAsset string, depthLevels int) string {
	var b strings.Builder

	// Header
	b.WriteString(ui.StrategyNameStyle.Render(fmt.Sprintf("ORDERBOOK - %s", selectedAsset)))
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  (Depth: %d levels)", depthLevels)))
	b.WriteString("\n\n")

	if orderbook == nil {
		b.WriteString(ui.SubtitleStyle.Render("No orderbook data available"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[D] Toggle Depth"))
		return b.String()
	}

	// Calculate max quantity for bar scaling
	maxQty := calculateMaxQuantity(orderbook, depthLevels)
	if maxQty == 0 {
		maxQty = 1
	}

	barWidth := 30

	// Asks (reversed - lowest ask at bottom)
	b.WriteString(askStyle.Render("                              ASKS"))
	b.WriteString("\n")

	asksToShow := depthLevels
	if asksToShow > len(orderbook.Asks) {
		asksToShow = len(orderbook.Asks)
	}

	// Show asks in reverse order (highest first)
	for i := asksToShow - 1; i >= 0; i-- {
		level := orderbook.Asks[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		bar := renderDepthBar(qty, maxQty, barWidth)
		row := fmt.Sprintf("  %12.2f  %s  %8.4f", price, askStyle.Render(bar), qty)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Spread
	if len(orderbook.Asks) > 0 && len(orderbook.Bids) > 0 {
		bestAsk, _ := orderbook.Asks[0].Price.Float64()
		bestBid, _ := orderbook.Bids[0].Price.Float64()
		spread := bestAsk - bestBid
		spreadPct := (spread / bestBid) * 100

		spreadLine := fmt.Sprintf("  ──────────── SPREAD: $%.2f (%.3f%%) ────────────", spread, spreadPct)
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(spreadLine))
		b.WriteString("\n")
	}

	// Bids
	bidsToShow := depthLevels
	if bidsToShow > len(orderbook.Bids) {
		bidsToShow = len(orderbook.Bids)
	}

	for i := 0; i < bidsToShow; i++ {
		level := orderbook.Bids[i]
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
	if len(orderbook.Asks) > 0 && len(orderbook.Bids) > 0 {
		bestAsk, _ := orderbook.Asks[0].Price.Float64()
		bestBid, _ := orderbook.Bids[0].Price.Float64()
		midPrice := (bestAsk + bestBid) / 2

		b.WriteString(fmt.Sprintf("Mid: $%.2f   Bid: $%.2f   Ask: $%.2f",
			midPrice, bestBid, bestAsk))
	}

	return b.String()
}

func calculateMaxQuantity(orderbook *connector.OrderBook, depthLevels int) float64 {
	maxQty := 0.0
	for i := 0; i < depthLevels && i < len(orderbook.Asks); i++ {
		qty, _ := orderbook.Asks[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	for i := 0; i < depthLevels && i < len(orderbook.Bids); i++ {
		qty, _ := orderbook.Bids[i].Quantity.Float64()
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
