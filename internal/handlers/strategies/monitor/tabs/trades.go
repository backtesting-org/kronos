package tabs

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/charmbracelet/lipgloss"
)

// RenderTrades renders the trades view
func RenderTrades(trades []connector.Trade) string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("RECENT TRADES"))
	b.WriteString("\n\n")

	if len(trades) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No trades yet"))
		return b.String()
	}

	// Table header
	b.WriteString(tableHeaderStyle.Render(fmt.Sprintf("  %-10s %-12s %-8s %-10s %-12s %-10s",
		"TIME", "SYMBOL", "SIDE", "QTY", "PRICE", "FEE")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 70)))
	b.WriteString("\n")

	for _, trade := range trades {
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
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Showing %d trades", len(trades))))

	return b.String()
}
