package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var nonInteractive bool

var rootCmd = &cobra.Command{
	Use:   "kronos",
	Short: "Kronos - Trading infrastructure platform",
	Long: `Kronos CLI - Beautiful backtesting and live trading infrastructure

Use Kronos to:
  • Configure backtests via YAML
  • Run backtests locally with deterministic simulation
  • Deploy strategies live
  • Analyze results

Examples:
  kronos                         Launch interactive menu
  kronos --non-interactive       Show traditional help
  kronos init my-project         Create a new project
  kronos backtest                Interactive backtest
  kronos live                    Interactive live trading`,
	Run: func(cmd *cobra.Command, args []string) {
		// If non-interactive flag or subcommand provided, show help
		if nonInteractive || len(args) > 0 {
			cmd.Help()
			return
		}

		// Launch main menu TUI
		if err := runMainMenu(cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Kronos CLI v%s\n", version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Disable interactive mode")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(backtestCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(liveCmd)
	rootCmd.AddCommand(versionCmd)
}

// Main menu TUI
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

	title := titleStyle.Render("KRONOS CLI v" + version)

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

func runMainMenu(rootCmd *cobra.Command) error {
	m := mainMenuModel{
		choices: []string{
			"Start Live Trading",
			"Run Backtest",
			"Analyze Results",
			"Create New Project",
			"Show Help",
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	result := finalModel.(mainMenuModel)

	// If user quit without selecting, just exit
	if result.selected == "" {
		return nil
	}

	// Route to appropriate command based on selection
	switch result.selected {
	case "Start Live Trading":
		return runLive(liveCmd, []string{})
	case "Run Backtest":
		interactiveMode = true
		return runBacktest(backtestCmd, []string{})
	case "Analyze Results":
		return runAnalyze(analyzeCmd, []string{})
	case "Create New Project":
		fmt.Print("\nEnter project name: ")
		var projectName string
		fmt.Scanln(&projectName)
		if projectName == "" {
			fmt.Println("Project name required")
			return nil
		}
		return runInit(initCmd, []string{projectName})
	case "Show Help":
		return showHelp()
	}

	return nil
}

// Help screen TUI
type helpModel struct {
	scrollOffset   int
	viewportHeight int
	quitting       bool
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewportHeight = msg.Height - 6
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			m.scrollOffset++
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		PaddingTop(1)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	// Build content as lines for scrolling
	lines := []string{}
	lines = append(lines, titleStyle.Render("🚀 KRONOS CLI v"+version))
	lines = append(lines, mutedStyle.Render("Trading infrastructure platform"))
	lines = append(lines, "")
	lines = append(lines, sectionStyle.Render("📋 COMMANDS"))
	lines = append(lines, "")

	commands := []struct{ cmd, desc string }{
		{"kronos", "Launch interactive menu"},
		{"kronos init <name>", "Create a new trading project"},
		{"kronos backtest", "Run backtests interactively"},
		{"kronos live", "Deploy strategies to live trading"},
		{"kronos analyze", "Analyze backtest results"},
		{"kronos version", "Show version information"},
	}

	for _, c := range commands {
		lines = append(lines, "  "+commandStyle.Render(c.cmd))
		lines = append(lines, "    "+descStyle.Render(c.desc))
		lines = append(lines, "")
	}

	// Handle scrolling
	start := m.scrollOffset
	end := len(lines)
	if m.viewportHeight > 0 && start+m.viewportHeight < end {
		end = start + m.viewportHeight
	}
	if start > len(lines) {
		start = len(lines)
	}
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	var s string
	for _, line := range visibleLines {
		s += line + "\n"
	}

	// Scroll indicators
	if start > 0 {
		s = mutedStyle.Render("↑ Scroll up for more") + "\n" + s
	}
	if end < len(lines) {
		s += mutedStyle.Render("↓ Scroll down for more") + "\n"
	}

	s += "\n" + mutedStyle.Render("↑↓/jk Scroll  q/esc/enter Exit")

	return "\n" + s + "\n"
}

func showHelp() error {
	m := helpModel{viewportHeight: 20}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
