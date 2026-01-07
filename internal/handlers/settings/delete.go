package settings

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/router"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// DeleteConfirmModel represents the delete confirmation view using Huh
type DeleteConfirmModel struct {
	form          *huh.Form
	connectorName string
	config        config.Configuration
	router        router.Router
	confirmed     bool
	err           error
}

// NewDeleteConfirmView creates a new delete confirmation view with Huh
func NewDeleteConfirmView(
	config config.Configuration,
	r router.Router,
	connectorName string,
) tea.Model {
	m := &DeleteConfirmModel{
		config:        config,
		router:        r,
		connectorName: connectorName,
	}

	// Build the confirmation form
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("🗑️  Delete Connector").
				Description(fmt.Sprintf(
					"Are you sure you want to delete connector **%s**?\n\n"+
						"⚠️  This action cannot be undone.",
					connectorName,
				)),

			huh.NewConfirm().
				Title("Delete this connector?").
				Description("This will permanently remove the connector configuration.").
				Affirmative("Delete").
				Negative("Cancel").
				Value(&m.confirmed),
		),
	).WithTheme(huh.ThemeCharm())

	return m
}

func (m *DeleteConfirmModel) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return nil
}

func (m *DeleteConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle error state
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "esc" {
				return m, m.router.Back()
			}
		}
		return m, nil
	}

	// Handle Ctrl+C
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Update the form
	var cmd tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is complete
	if m.form.State == huh.StateCompleted {
		if m.confirmed {
			// User confirmed - delete the connector
			if err := m.config.RemoveConnector(m.connectorName); err != nil {
				m.err = err
				return m, nil
			}
		}
		// Go back to list (whether deleted or cancelled)
		return m, m.router.Back()
	}

	// Check if form was aborted
	if m.form.State == huh.StateAborted {
		return m, m.router.Back()
	}

	return m, cmd
}

func (m *DeleteConfirmModel) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"❌ Error\n\n%s\n\nPress 'q' or 'Esc' to go back.",
			m.err.Error(),
		)
	}

	if m.form == nil {
		return "Loading..."
	}

	return m.form.View()
}
