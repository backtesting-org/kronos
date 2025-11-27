package handlers

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/setup/scaffold"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen types for init flow
type InitScreen int

const (
	InitScreenStrategy InitScreen = iota
	InitScreenProjectName
)

// Strategy templates
type StrategyTemplate struct {
	Name        string
	DisplayName string
	Description string
	Icon        string
	SDKExample  string // Maps to SDK examples directory
}

// InitTUIModel represents the init flow TUI state
type InitTUIModel struct {
	screen            InitScreen
	cursor            int
	strategies        []StrategyTemplate
	selectedStrategy  *StrategyTemplate
	projectName       string
	projectNameInput  string
	err               error
	loadingStrategies bool
}

func NewInitTUIModel(strategies []StrategyTemplate) InitTUIModel {
	return InitTUIModel{
		screen:     InitScreenStrategy,
		strategies: strategies,
	}
}

func (m InitTUIModel) Init() tea.Cmd {
	return nil
}

func (m InitTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case InitScreenStrategy:
		return m.updateStrategySelection(msg)
	case InitScreenProjectName:
		return m.updateProjectName(msg)
	}
	return m, nil
}

func (m InitTUIModel) updateStrategySelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.strategies)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.strategies) > 0 {
				m.selectedStrategy = &m.strategies[m.cursor]
				m.screen = InitScreenProjectName
			}
			return m, nil
		}
	}
	return m, nil
}

func (m InitTUIModel) updateProjectName(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Go back to strategy selection
			m.screen = InitScreenStrategy
			m.projectNameInput = ""
			return m, nil
		case "enter":
			if m.projectNameInput == "" {
				m.err = fmt.Errorf("project name cannot be empty")
				return m, nil
			}
			// Convert spaces to underscores
			m.projectName = strings.ReplaceAll(m.projectNameInput, " ", "_")
			return m, tea.Quit
		case "backspace":
			if len(m.projectNameInput) > 0 {
				m.projectNameInput = m.projectNameInput[:len(m.projectNameInput)-1]
			}
		default:
			// Only allow alphanumeric, spaces, underscores, and hyphens
			if len(msg.String()) == 1 {
				char := msg.String()[0]
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9') || char == '_' || char == '-' || char == ' ' {
					m.projectNameInput += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m InitTUIModel) View() string {
	switch m.screen {
	case InitScreenStrategy:
		return m.viewStrategySelection()
	case InitScreenProjectName:
		return m.viewProjectName()
	}
	return ""
}

func (m InitTUIModel) viewStrategySelection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(2, 4).
		Width(70)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		PaddingLeft(4).
		Width(60)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	title := titleStyle.Render("🆕 CREATE NEW PROJECT")

	var s string
	s += "\n" + title + "\n\n"

	if len(m.strategies) == 0 {
		s += mutedStyle.Render("No strategies available. Check SDK connection.") + "\n\n"
		s += mutedStyle.Render("q Quit")
		return boxStyle.Render(s)
	}

	s += mutedStyle.Render("Select a strategy template to get started:") + "\n\n"

	for i, strategy := range m.strategies {
		cursor := "  "
		icon := strategy.Icon
		if icon == "" {
			icon = "🎯"
		}

		if m.cursor == i {
			cursor = "▶ "
			s += selectedStyle.Render(cursor+icon+" "+strategy.DisplayName) + "\n"
			s += descStyle.Render(strategy.Description) + "\n"
		} else {
			s += itemStyle.Render(cursor+icon+" "+strategy.DisplayName) + "\n"
		}
		if i < len(m.strategies)-1 {
			s += "\n"
		}
	}

	s += "\n\n" + mutedStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")

	return boxStyle.Render(s)
}

func (m InitTUIModel) viewProjectName() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(2, 4).
		Width(70)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB"))

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444"))

	title := titleStyle.Render("🆕 CREATE NEW PROJECT")

	var s string
	s += "\n" + title + "\n\n"
	s += labelStyle.Render(fmt.Sprintf("Selected Strategy: %s", m.selectedStrategy.DisplayName)) + "\n\n"
	s += mutedStyle.Render("Enter a name for your project:") + "\n"
	s += mutedStyle.Render("(Spaces will be converted to underscores)") + "\n\n"

	s += labelStyle.Render("Project Name: ") + inputStyle.Render(m.projectNameInput+"_") + "\n\n"

	if m.err != nil {
		s += errorStyle.Render("✗ "+m.err.Error()) + "\n\n"
	}

	s += mutedStyle.Render("↵ Create  ⎋ Back  ^C Cancel")

	return boxStyle.Render(s)
}

// RunInitTUI runs the init TUI flow and returns the selected strategy and project name
func RunInitTUI() (strategy string, projectName string, err error) {
	// Load available strategies from SDK
	strategies, err := LoadStrategies()
	if err != nil {
		return "", "", fmt.Errorf("failed to load strategies: %w", err)
	}

	m := NewInitTUIModel(strategies)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", "", err
	}

	result := finalModel.(InitTUIModel)
	if result.selectedStrategy == nil || result.projectName == "" {
		return "", "", fmt.Errorf("initialization cancelled")
	}

	return result.selectedStrategy.SDKExample, result.projectName, nil
}

// LoadStrategies loads strategy metadata from SDK or uses fallback
func LoadStrategies() ([]StrategyTemplate, error) {
	// Try to fetch from SDK
	metadata, err := fetchFromSDK()
	if err == nil && len(metadata) > 0 {
		return metadata, nil
	}

	// Fallback to hardcoded list if SDK fetch fails
	return getFallbackStrategies(), nil
}

// fetchFromSDK attempts to fetch strategy metadata from SDK repository
func fetchFromSDK() ([]StrategyTemplate, error) {
	metadata, err := scaffold.FetchAvailableStrategies()
	if err != nil {
		return nil, err
	}

	// Convert to StrategyTemplate
	templates := make([]StrategyTemplate, 0, len(metadata))
	for _, m := range metadata {
		icon := m.Icon
		if icon == "" {
			icon = scaffold.GetDefaultIcon(m.Type)
		}

		templates = append(templates, StrategyTemplate{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Description: m.Description,
			Icon:        icon,
			SDKExample:  m.SDKExample,
		})
	}

	return templates, nil
}

// getFallbackStrategies returns a hardcoded list of strategies as fallback
func getFallbackStrategies() []StrategyTemplate {
	return []StrategyTemplate{
		{
			Name:        "mean_reversion",
			DisplayName: "Mean Reversion Strategy",
			Description: "Bollinger Bands mean reversion with RSI confirmation",
			Icon:        "📉",
			SDKExample:  "mean_reversion",
		},
	}
}
