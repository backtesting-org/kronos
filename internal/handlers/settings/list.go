package settings

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/router"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

// ConnectorListModel represents the settings list view
type ConnectorListModel struct {
	configured         []config.Connector // Already configured connectors
	available          []string           // Available but not configured
	cursor             int
	inAvailableSection bool // true if cursor is in "add new" section
	config             config.Configuration
	connectorSvc       config.ConnectorService
	router             router.Router
	formFactory        ConnectorFormViewFactory
	deleteFactory      DeleteConfirmViewFactory
	err                error
	successMsg         string
}

// NewSettingsListView creates a new settings list view
func NewSettingsListView(
	cfg config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	formFactory ConnectorFormViewFactory,
	deleteFactory DeleteConfirmViewFactory,
) tea.Model {
	return &ConnectorListModel{
		config:        cfg,
		connectorSvc:  connectorSvc,
		router:        r,
		formFactory:   formFactory,
		deleteFactory: deleteFactory,
		configured:    []config.Connector{},
		available:     []string{},
	}
}

func (m *ConnectorListModel) Init() tea.Cmd {
	// Load configured connectors
	connectorList, err := m.config.GetConnectors()
	if err != nil {
		m.err = err
		return nil
	}
	m.configured = connectorList

	// Get available connectors from SDK
	allAvailable := m.connectorSvc.GetAvailableConnectorNames()

	// Filter out already configured ones
	configuredMap := make(map[string]bool)
	for _, c := range m.configured {
		configuredMap[c.Name] = true
	}

	m.available = []string{}
	for _, name := range allAvailable {
		if !configuredMap[name] {
			m.available = append(m.available, name)
		}
	}

	return nil
}

func (m *ConnectorListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Navigate back to main menu
			return m, m.router.Back()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Update section tracking
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "down", "j":
			totalItems := len(m.configured) + len(m.available)
			if m.cursor < totalItems-1 {
				m.cursor++
				// Update section tracking
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "enter":
			if m.inAvailableSection {
				// Add new connector from available list
				availableIndex := m.cursor - len(m.configured)
				if availableIndex >= 0 && availableIndex < len(m.available) {
					selectedExchange := m.available[availableIndex]
					createView := m.formFactory(selectedExchange, false)
					return m, bubblon.Open(createView)
				}
			} else {
				// Edit configured connector
				if m.cursor < len(m.configured) {
					selectedConnectorName := m.configured[m.cursor].Name
					editView := m.formFactory(selectedConnectorName, true)
					return m, bubblon.Open(editView)
				}
			}
		case "d":
			// Delete only works on configured connectors
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				selectedConnectorName := m.configured[m.cursor].Name
				deleteView := m.deleteFactory(selectedConnectorName)
				return m, bubblon.Open(deleteView)
			}
		case " ":
			// Toggle only works on configured connectors
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				connectorName := m.configured[m.cursor].Name
				newState := !m.configured[m.cursor].Enabled
				if err := m.config.EnableConnector(connectorName, newState); err != nil {
					m.err = err
				} else {
					// Reload - calls Init() which refreshes both lists
					m.Init()
					m.successMsg = "Connector updated"
				}
			}
		}
	}
	return m, nil
}

func (m *ConnectorListModel) View() string {
	var content strings.Builder

	// Title
	title := ui.TitleStyle.Render("⚙️  Connector Configuration")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Error message if any
	if m.err != nil {
		errorBox := lipgloss.NewStyle().
			Foreground(ui.ColorDanger).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorDanger).
			Padding(0, 2).
			Width(68).
			Render("❌ " + m.err.Error())
		content.WriteString(errorBox)
		content.WriteString("\n\n")
	}

	// Success message if any
	if m.successMsg != "" {
		successMsg := lipgloss.NewStyle().
			Foreground(ui.ColorSuccess).
			Bold(true).
			Render("✓ " + m.successMsg)
		content.WriteString(successMsg)
		content.WriteString("\n\n")
	}

	// Section 1: Configured Connectors
	sectionHeader := lipgloss.NewStyle().
		Foreground(ui.ColorSecondary).
		Bold(true).
		Render("📋 CONFIGURED CONNECTORS")
	content.WriteString(sectionHeader)
	content.WriteString("\n\n")

	if len(m.configured) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Italic(true).
			Render("   No connectors configured yet")
		content.WriteString(emptyMsg)
		content.WriteString("\n")
	} else {
		for i, conn := range m.configured {
			isSelected := m.cursor == i && !m.inAvailableSection
			content.WriteString(m.renderConfiguredConnector(conn, isSelected))
		}
	}

	content.WriteString("\n")

	// Section 2: Add New Connector
	addHeader := lipgloss.NewStyle().
		Foreground(ui.ColorSecondary).
		Bold(true).
		Render("➕ ADD NEW CONNECTOR")
	content.WriteString(addHeader)
	content.WriteString("\n\n")

	if len(m.available) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Italic(true).
			Render("   All available connectors are configured")
		content.WriteString(emptyMsg)
		content.WriteString("\n")
	} else {
		for i, name := range m.available {
			globalIndex := len(m.configured) + i
			isSelected := m.cursor == globalIndex
			content.WriteString(m.renderAvailableConnector(name, isSelected))
		}
	}

	content.WriteString("\n")

	// Help text (context-aware)
	helpText := m.getHelpText()
	help := ui.HelpStyle.Render(helpText)
	content.WriteString(help)

	return content.String()
}

func (m *ConnectorListModel) renderConfiguredConnector(conn config.Connector, selected bool) string {
	var itemStyle lipgloss.Style
	var nameStyle lipgloss.Style

	if selected {
		itemStyle = ui.StrategyItemSelectedStyle
		nameStyle = ui.StrategyNameSelectedStyle
	} else {
		itemStyle = ui.StrategyItemStyle
		nameStyle = ui.StrategyNameStyle
	}

	// Build content
	var content strings.Builder

	// Name with network badge
	name := nameStyle.Render(conn.Name)
	if conn.Network != "" {
		networkColor := ui.ColorSecondary
		if conn.Network == "testnet" {
			networkColor = ui.ColorWarning
		}
		networkBadge := lipgloss.NewStyle().
			Foreground(networkColor).
			Render(" [" + conn.Network + "]")
		name += networkBadge
	}
	content.WriteString(name)
	content.WriteString("\n")

	// Status indicator
	var statusStyle lipgloss.Style
	var statusText string
	if conn.Enabled {
		statusStyle = ui.StatusReadyStyle
		statusText = "● ENABLED"
	} else {
		statusStyle = lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Bold(true)
		statusText = "○ DISABLED"
	}
	status := statusStyle.Render(statusText)
	content.WriteString(status)

	return itemStyle.Render(content.String())
}

func (m *ConnectorListModel) renderAvailableConnector(name string, selected bool) string {
	cursor := "  "
	style := lipgloss.NewStyle().Foreground(ui.ColorMuted)

	if selected {
		cursor = lipgloss.NewStyle().
			Foreground(ui.ColorPrimary).
			Bold(true).
			Render("▶ ")
		style = lipgloss.NewStyle().
			Foreground(ui.ColorPrimary).
			Bold(true)
	}

	bullet := lipgloss.NewStyle().
		Foreground(ui.ColorSecondary).
		Render("• ")

	line := cursor + bullet + style.Render(name) + "\n"
	return line
}

func (m *ConnectorListModel) getHelpText() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true)

	if m.inAvailableSection {
		return fmt.Sprintf(
			"%s/%s Navigate  %s Add Connector  %s Back",
			keyStyle.Render("↑"),
			keyStyle.Render("↓"),
			keyStyle.Render("Enter"),
			keyStyle.Render("q"),
		)
	}

	return fmt.Sprintf(
		"%s/%s Navigate  %s Edit  %s Delete  %s Toggle  %s Back",
		keyStyle.Render("↑"),
		keyStyle.Render("↓"),
		keyStyle.Render("Enter"),
		keyStyle.Render("d"),
		keyStyle.Render("Space"),
		keyStyle.Render("q"),
	)
}
