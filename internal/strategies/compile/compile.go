package compile

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type CompileModel interface {
	tea.Model
	SetStrategy(strategy *strategy.Strategy)
	Done() bool
}

type compileModel struct {
	strategy       *strategy.Strategy
	compileService shared.CompileService
	done           bool
	err            error
	output         string
}

// NewCompileModel creates a compile view with all dependencies
func NewCompileModel(compileService shared.CompileService) CompileModel {
	return compileModel{
		strategy:       nil,
		compileService: compileService,
		done:           false,
	}
}

func (m compileModel) SetStrategy(strategy *strategy.Strategy) {
	m.strategy = strategy
}

func (m compileModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Run compile in background
		err := m.compileService.CompileStrategy(m.strategy.Path)
		return CompileFinishedMsg{Err: err}
	}
}

func (m compileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CompileFinishedMsg:
		m.done = true
		m.err = msg.Err
		return m, nil
	case tea.KeyMsg:
		// Don't allow interaction during compile
		return m, nil
	}
	return m, nil
}

func (m compileModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"
	content += ui.SubtitleStyle.Render("Compiling...") + "\n\n"
	content += "⏳ Building plugin...\n"
	return ui.BoxStyle.Render(content)
}

// Done returns whether compilation is complete
func (m compileModel) Done() bool {
	return m.done
}

// GetStrategy returns the strategy being compiled
func (m compileModel) GetStrategy() *strategy.Strategy {
	return m.strategy
}

// GetError returns any compilation error
func (m compileModel) GetError() error {
	return m.err
}

// CompileFinishedMsg is sent when compilation completes
type CompileFinishedMsg struct {
	Err error
}
