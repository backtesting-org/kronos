package handlers

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// mainMenuModel represents the main menu TUI
type mainMenuModel struct {
	choices  []string
	cursor   int
	selected string
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mainMenuModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1).
		Align(lipgloss.Center)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(2, 4).
		Width(50)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true).
		PaddingLeft(0)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	title := titleStyle.Render("KRONOS CLI v0.1.0")

	var s string
	s += "\n" + title + "\n\n"
	s += mutedStyle.Render("What would you like to do?") + "\n\n"

	icons := []string{"🚀", "📊", "📈", "🆕", "ℹ️"}

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
			s += selectedStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		} else {
			s += itemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		}
	}

	s += "\n" + mutedStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")

	return boxStyle.Render(s)
}

func (h *RootHandler) handleCreateProject(cmd *cobra.Command) error {
	// Build a styled input prompt
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF"))

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	fmt.Println()
	fmt.Println(titleStyle.Render("🆕 CREATE NEW PROJECT"))
	fmt.Println()
	fmt.Print(promptStyle.Render("Project name: "))

	var projectName string
	fmt.Scanln(&projectName)

	if projectName == "" {
		fmt.Println(mutedStyle.Render("✗ Project name required"))
		return nil
	}

	return h.initHandler.Handle(cmd, []string{projectName})
}
