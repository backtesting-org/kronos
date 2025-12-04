package tabs

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

// Shared styles for tab views
var (
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(ui.ColorMuted).
				Bold(true)

	profitStyle = lipgloss.NewStyle().
			Foreground(ui.ColorSuccess).
			Bold(true)

	lossStyle = lipgloss.NewStyle().
			Foreground(ui.ColorDanger).
			Bold(true)

	neutralStyle = lipgloss.NewStyle().
			Foreground(ui.ColorMuted)
)

// FormatPnL formats a PnL value with appropriate styling
func FormatPnL(value float64) string {
	if value > 0 {
		return profitStyle.Render(fmt.Sprintf("+$%.2f", value))
	} else if value < 0 {
		return lossStyle.Render(fmt.Sprintf("-$%.2f", -value))
	}
	return neutralStyle.Render("$0.00")
}
