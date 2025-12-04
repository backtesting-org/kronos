package tabs

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/charmbracelet/lipgloss"
)

// RenderOverview renders the overview tab with summary panels
func RenderOverview(pnl *monitoring.PnLView, metrics *monitoring.StrategyMetrics) string {
	var b strings.Builder

	// PnL Summary Panel
	pnlContent := RenderPnLSummary(pnl)
	pnlPanel := ui.BoxStyle.Width(35).Render(pnlContent)

	// Quick Stats Panel
	statsContent := RenderQuickStats(metrics)
	statsPanel := ui.BoxStyle.Width(35).Render(statsContent)

	// Side by side
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, pnlPanel, "  ", statsPanel))

	return b.String()
}

// RenderQuickStats renders the quick stats panel
func RenderQuickStats(metrics *monitoring.StrategyMetrics) string {
	var b strings.Builder
	b.WriteString(ui.StrategyNameStyle.Render("QUICK STATS"))
	b.WriteString("\n\n")

	if metrics == nil {
		b.WriteString(ui.SubtitleStyle.Render("No data"))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Signals Generated:  %d\n", metrics.SignalsGenerated))
	b.WriteString(fmt.Sprintf("Signals Executed:   %d\n", metrics.SignalsExecuted))

	successRate := 0.0
	if metrics.SignalsGenerated > 0 {
		successRate = float64(metrics.SignalsExecuted) / float64(metrics.SignalsGenerated) * 100
	}
	b.WriteString(fmt.Sprintf("Success Rate:       %.0f%%\n", successRate))
	b.WriteString(fmt.Sprintf("Avg Latency:        %v", metrics.AverageLatency))

	return b.String()
}
