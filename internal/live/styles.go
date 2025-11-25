package live

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00D9FF") // Cyan
	ColorSecondary = lipgloss.Color("#7C3AED") // Purple
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Orange
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#6B7280") // Gray

	// Background colors
	ColorBgDark     = lipgloss.Color("#1F2937")
	ColorBgMedium   = lipgloss.Color("#374151")
	ColorBgLight    = lipgloss.Color("#4B5563")
	ColorBgSelected = lipgloss.Color("#1E293B")
)

// Style definitions
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(1, 2).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// Strategy item styles
	StrategyItemStyle = lipgloss.NewStyle().
				Padding(1, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBgLight).
				Width(70)

	StrategyItemSelectedStyle = lipgloss.NewStyle().
					Padding(1, 2).
					MarginBottom(1).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ColorPrimary).
					Background(ColorBgSelected).
					Width(70)

	// Strategy name
	StrategyNameStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	StrategyNameSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorSuccess).
					Bold(true)

	// Strategy description
	StrategyDescStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)

	// Strategy metadata
	StrategyMetaStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				MarginTop(1)

	// Status indicators
	StatusReadyStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	StatusDangerStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				MarginTop(1)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 2).
			MarginTop(1)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(70)

	// Confirmation styles
	ConfirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ColorDanger).
			Padding(2, 4).
			Width(70)

	ConfirmTitleStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true).
				Align(lipgloss.Center)

	ConfirmFieldStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	ConfirmValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	ConfirmWarningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Align(lipgloss.Center).
				MarginTop(1).
				MarginBottom(1)

	// Input style
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
)

// GetStatusIndicator returns a styled status indicator
func GetStatusIndicator(status StrategyStatus) string {
	switch status {
	case StatusReady:
		return StatusReadyStyle.Render("● READY")
	case StatusRunning:
		return StatusRunningStyle.Render("● RUNNING")
	case StatusStopped:
		return StatusDangerStyle.Render("● STOPPED")
	case StatusError:
		return StatusDangerStyle.Render("● ERROR")
	default:
		return StatusReadyStyle.Render("● READY")
	}
}

// GetModeIndicator returns a styled mode indicator
func GetModeIndicator(dryRun bool) string {
	if dryRun {
		return StatusReadyStyle.Render("📝 PAPER TRADING")
	}
	return StatusDangerStyle.Render("🔴 LIVE TRADING")
}
