package settings

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/settings/connectors"
	"github.com/backtesting-org/kronos-cli/internal/router"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

// ConnectorListModel represents the settings list view
type ConnectorListModel struct {
	connectors    []settings.Connector
	cursor        int
	config        settings.Configuration
	connectorSvc  connectors.ConnectorService
	router        router.Router
	formFactory   ConnectorFormViewFactory
	deleteFactory DeleteConfirmViewFactory
	err           error
	successMsg    string
}

// NewSettingsListView creates a new settings list view
func NewSettingsListView(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	formFactory ConnectorFormViewFactory,
	deleteFactory DeleteConfirmViewFactory,
) tea.Model {
	return &ConnectorListModel{
		config:        config,
		connectorSvc:  connectorSvc,
		router:        r,
		formFactory:   formFactory,
		deleteFactory: deleteFactory,
		connectors:    []settings.Connector{},
	}
}

func (m *ConnectorListModel) Init() tea.Cmd {
	// Load connectors
	connectorList, err := m.config.GetConnectors()
	if err != nil {
		m.err = err
		return nil
	}
	m.connectors = connectorList
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
			}
		case "down", "j":
			if m.cursor < len(m.connectors)-1 {
				m.cursor++
			}
		case "enter":
			// Edit selected connector
			if len(m.connectors) > 0 {
				selectedConnectorName := m.connectors[m.cursor].Name
				editView := m.formFactory(selectedConnectorName, true)
				return m, bubblon.Open(editView)
			}
		case "n":
			// Create new connector
			createView := m.formFactory("", false)
			return m, bubblon.Open(createView)
		case "d":
			// Delete selected connector
			if len(m.connectors) > 0 {
				selectedConnectorName := m.connectors[m.cursor].Name
				deleteView := m.deleteFactory(selectedConnectorName)
				return m, bubblon.Open(deleteView)
			}
		case " ":
			// Toggle enabled/disabled
			if len(m.connectors) > 0 {
				connectorName := m.connectors[m.cursor].Name
				newState := !m.connectors[m.cursor].Enabled
				if err := m.config.EnableConnector(connectorName, newState); err != nil {
					m.err = err
				} else {
					// Reload connectors
					connectorList, err := m.config.GetConnectors()
					if err != nil {
						m.err = err
					} else {
						m.connectors = connectorList
						m.successMsg = "Connector updated"
					}
				}
			}
		}
	}
	return m, nil
}

func (m *ConnectorListModel) View() string {
	s := "⚙️  Connector Configuration\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	if m.err != nil {
		s += "❌ Error: " + m.err.Error() + "\n\n"
	}

	if m.successMsg != "" {
		s += "✅ " + m.successMsg + "\n\n"
	}

	if len(m.connectors) == 0 {
		s += "No connectors configured.\n"
		s += "Press 'n' to create a new connector.\n"
	} else {
		for i, conn := range m.connectors {
			cursor := "  "
			if m.cursor == i {
				cursor = "▶ "
			}

			status := "✗"
			statusText := "Disabled"
			if conn.Enabled {
				status = "✓"
				statusText = "Enabled"
			}

			network := ""
			if conn.Network != "" {
				network = " [" + conn.Network + "]"
			}

			s += cursor + status + "  " + conn.Name + network + "     " + statusText + "\n"
		}
	}

	s += "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	s += "↑↓ Navigate  ↵ Edit  n New  d Delete  Space Toggle  q Back\n"

	return s
}
