package tabs

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/charmbracelet/lipgloss"
)

// RenderPositions renders the positions view
func RenderPositions(positions *strategy.StrategyExecution) string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("POSITIONS"))
	b.WriteString("\n\n")

	if positions == nil || (len(positions.Orders) == 0 && len(positions.Trades) == 0) {
		b.WriteString(ui.SubtitleStyle.Render("No active positions"))
		return b.String()
	}

	// Show orders
	if len(positions.Orders) > 0 {
		b.WriteString(tableHeaderStyle.Render(fmt.Sprintf("  %-12s %-8s %-10s %-12s %-12s", "SYMBOL", "SIDE", "QTY", "PRICE", "STATUS")))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 60)))
		b.WriteString("\n")

		for _, order := range positions.Orders {
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
