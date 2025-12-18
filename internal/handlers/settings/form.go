package settings

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/settings/connectors"
	"github.com/backtesting-org/kronos-cli/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// ConnectorFormModel represents the create/edit connector form
type ConnectorFormModel struct {
	connector    settings.Connector
	config       settings.Configuration
	connectorSvc connectors.ConnectorService
	router       router.Router
	isEditMode   bool
	originalName string

	// Field editing state
	focusedField int
	editing      bool
	editBuffer   string
	err          error
	successMsg   string
}

// NewConnectorFormView creates a new connector form view
func NewConnectorFormView(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	connectorName string,
	isEdit bool,
) tea.Model {
	m := &ConnectorFormModel{
		config:       config,
		connectorSvc: connectorSvc,
		router:       r,
		isEditMode:   isEdit,
		originalName: connectorName,
	}

	if isEdit && connectorName != "" {
		// Load existing connector
		connectorList, err := config.GetConnectors()
		if err != nil {
			m.err = err
			return m
		}

		for _, conn := range connectorList {
			if conn.Name == connectorName {
				m.connector = conn
				break
			}
		}

		if m.connector.Name == "" {
			m.err = fmt.Errorf("connector '%s' not found", connectorName)
		}
	} else {
		// Initialize empty connector for creation
		m.connector = settings.Connector{
			Enabled:     true,
			Credentials: make(map[string]string),
		}
	}

	return m
}

func (m *ConnectorFormModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectorFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingKeys(msg)
		}
		return m.handleNavigationKeys(msg)
	}
	return m, nil
}

func (m *ConnectorFormModel) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, m.router.Back()
	case "up", "k":
		if m.focusedField > 0 {
			m.focusedField--
		}
	case "down", "j":
		maxFields := 3 + len(m.connector.Credentials)
		if m.focusedField < maxFields {
			m.focusedField++
		}
	case "enter":
		// Start editing the focused field
		m.editing = true
		m.editBuffer = m.getFieldValue(m.focusedField)
	case "s":
		// Save connector
		return m.saveConnector()
	}
	return m, nil
}

func (m *ConnectorFormModel) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.editBuffer = ""
	case "enter":
		m.setFieldValue(m.focusedField, m.editBuffer)
		m.editing = false
		m.editBuffer = ""
	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.editBuffer += msg.String()
		}
	}
	return m, nil
}

func (m *ConnectorFormModel) getFieldValue(field int) string {
	switch field {
	case 0:
		return m.connector.Name
	case 1:
		return m.connector.Network
	case 2:
		if m.connector.Enabled {
			return "true"
		}
		return "false"
	default:
		// Credentials fields
		credIndex := field - 3
		credKeys := m.getCredentialKeys()
		if credIndex >= 0 && credIndex < len(credKeys) {
			return m.connector.Credentials[credKeys[credIndex]]
		}
	}
	return ""
}

func (m *ConnectorFormModel) setFieldValue(field int, value string) {
	switch field {
	case 0:
		m.connector.Name = value
	case 1:
		m.connector.Network = value
	case 2:
		m.connector.Enabled = value == "true"
	default:
		// Credentials fields
		credIndex := field - 3
		credKeys := m.getCredentialKeys()
		if credIndex >= 0 && credIndex < len(credKeys) {
			m.connector.Credentials[credKeys[credIndex]] = value
		}
	}
}

func (m *ConnectorFormModel) getCredentialKeys() []string {
	keys := make([]string, 0, len(m.connector.Credentials))
	for key := range m.connector.Credentials {
		keys = append(keys, key)
	}
	return keys
}

func (m *ConnectorFormModel) saveConnector() (tea.Model, tea.Cmd) {
	var err error
	if m.isEditMode {
		err = m.config.UpdateConnector(m.connector)
	} else {
		err = m.config.AddConnector(m.connector)
	}

	if err != nil {
		m.err = err
		return m, nil
	}

	m.successMsg = "Connector saved successfully"
	return m, m.router.Back()
}

func (m *ConnectorFormModel) View() string {
	title := "📝 Create New Connector"
	if m.isEditMode {
		title = "✏️  Edit Connector: " + m.originalName
	}

	s := title + "\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	if m.err != nil {
		s += "❌ Error: " + m.err.Error() + "\n\n"
	}

	if m.successMsg != "" {
		s += "✅ " + m.successMsg + "\n\n"
	}

	// Display fields
	s += m.renderField(0, "Name", m.connector.Name)
	s += m.renderField(1, "Network", m.connector.Network)
	s += m.renderField(2, "Enabled", fmt.Sprintf("%t", m.connector.Enabled))

	s += "\nCredentials:\n"
	credKeys := m.getCredentialKeys()
	for i, key := range credKeys {
		value := m.connector.Credentials[key]
		// Mask sensitive values
		maskedValue := "********"
		if len(value) < 8 {
			maskedValue = value
		}
		s += m.renderField(3+i, key, maskedValue)
	}

	s += "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	if m.editing {
		s += "Editing: " + m.editBuffer + "\n"
		s += "↵ Confirm  Esc Cancel\n"
	} else {
		s += "↑↓ Navigate  ↵ Edit  s Save  q Back\n"
	}

	return s
}

func (m *ConnectorFormModel) renderField(index int, label, value string) string {
	cursor := "  "
	if m.focusedField == index {
		cursor = "▶ "
	}

	displayValue := value
	if m.editing && m.focusedField == index {
		displayValue = m.editBuffer + "█"
	}

	return fmt.Sprintf("%s%-15s: %s\n", cursor, label, displayValue)
}
