package tabs

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/charmbracelet/lipgloss"
)

// RenderPnL renders the detailed PnL breakdown view
func RenderPnL(pnl *monitoring.PnLView) string {
	var b strings.Builder

	b.WriteString(ui.StrategyNameStyle.Render("PROFIT & LOSS BREAKDOWN"))
	b.WriteString("\n\n")

	if pnl == nil {
		b.WriteString(ui.SubtitleStyle.Render("No PnL data available"))
		return b.String()
	}

	realized, _ := pnl.RealizedPnL.Float64()
	unrealized, _ := pnl.UnrealizedPnL.Float64()
	total, _ := pnl.TotalPnL.Float64()
	fees, _ := pnl.TotalFees.Float64()

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

// RenderPnLSummary renders a compact PnL summary for overview
func RenderPnLSummary(pnl *monitoring.PnLView) string {
	var b strings.Builder
	b.WriteString(ui.StrategyNameStyle.Render("PNL SUMMARY"))
	b.WriteString("\n\n")

	if pnl == nil {
		b.WriteString(ui.SubtitleStyle.Render("No data"))
		return b.String()
	}

	realized, _ := pnl.RealizedPnL.Float64()
	unrealized, _ := pnl.UnrealizedPnL.Float64()
	total, _ := pnl.TotalPnL.Float64()
	fees, _ := pnl.TotalFees.Float64()

	b.WriteString(fmt.Sprintf("Realized:    %s\n", FormatPnL(realized)))
	b.WriteString(fmt.Sprintf("Unrealized:  %s\n", FormatPnL(unrealized)))
	b.WriteString(fmt.Sprintf("Total:       %s\n", FormatPnL(total)))
	b.WriteString(fmt.Sprintf("Fees:        %s", lossStyle.Render(fmt.Sprintf("-$%.2f", fees))))

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
