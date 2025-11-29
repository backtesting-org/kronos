package compile

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type CompileModel struct {
	strategy       *strategy.Strategy
	compileService shared.CompileService
	done           bool
	err            error
	output         string
}

// NewCompileModel creates a compile view with all dependencies
func NewCompileModel(compileService shared.CompileService, strat *strategy.Strategy) tea.Model {
	return CompileModel{
		strategy:       strat,
		compileService: compileService,
		done:           false,
	}
}

func (m CompileModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Run compile in background
		err := m.compileService.CompileStrategy(m.strategy.Path)
		return CompileFinishedMsg{Err: err}
	}
}

func (m CompileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m CompileModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"
	content += ui.SubtitleStyle.Render("Compiling...") + "\n\n"
	content += "⏳ Building plugin...\n"
	return ui.BoxStyle.Render(content)
}

// Done returns whether compilation is complete
func (m CompileModel) Done() bool {
	return m.done
}

// GetStrategy returns the strategy being compiled
func (m CompileModel) GetStrategy() *strategy.Strategy {
	return m.strategy
}

// GetError returns any compilation error
func (m CompileModel) GetError() error {
	return m.err
}

// CompileFinishedMsg is sent when compilation completes
type CompileFinishedMsg struct {
	Err error
}

// CompileResultModel shows compile results (COMPILE-RESULT screen)
type ResultModel struct {
	strategy     *strategy.Strategy
	err          error
	backToDetail bool
}

// NewResultModel creates a result model - called directly when compile finishes
func NewResultModel(strat *strategy.Strategy, err error) ResultModel {
	return ResultModel{
		strategy:     strat,
		err:          err,
		backToDetail: false,
	}
}

func (m ResultModel) Init() tea.Cmd {
	return nil
}

func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			m.backToDetail = true
		}
	}
	return m, nil
}

func (m ResultModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"

	if m.err == nil {
		content += ui.StatusReadyStyle.Render("✅ Compilation Successful") + "\n\n"
		content += "Strategy has been compiled to .so plugin\n"
		content += "Ready for backtest or live trading\n"
	} else {
		content += ui.StatusErrorStyle.Render("❌ Compilation Failed") + "\n\n"
		content += ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	content += "\n" + ui.SubtitleStyle.Render("Press Enter or q to go back")

	return ui.BoxStyle.Render(content)
}

// ShouldBackToDetail returns whether to navigate back to detail
func (m ResultModel) ShouldBackToDetail() bool {
	return m.backToDetail
}

// Reset clears the back flag
func (m ResultModel) Reset() {
	m.backToDetail = false
}
