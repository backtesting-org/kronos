package ui

import (
	"fmt"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/pterm/pterm"
	"github.com/schollz/progressbar/v3"
)

// ShowBanner displays the ASCII art "KRONOS" banner
func ShowBanner() {
	myFigure := figure.NewFigure("KRONOS", "big", true)
	pterm.DefaultCenter.Println(pterm.Cyan(myFigure.String()))
	fmt.Println()
}

// Success prints a success message with checkmark
func Success(message string) {
	pterm.Success.Println(message)
}

// Info prints an info message
func Info(message string) {
	pterm.Info.Println(message)
}

// Warning prints a warning message
func Warning(message string) {
	pterm.Warning.Println(message)
}

// Error prints an error message
func Error(message string) {
	pterm.Error.Println(message)
}

// Section prints a section header
func Section(title string) {
	pterm.DefaultSection.Println(title)
}

// BacktestResults represents the results of a backtest
type BacktestResults struct {
	TotalPnL     float64
	WinRate      float64
	TotalTrades  int
	AvgTradePnL  float64
	SharpeRatio  float64
	MaxDrawdown  float64
	ProfitFactor float64
	Duration     time.Duration
	ResultsFile  string
}

// DisplayResults shows backtest results in a beautiful table
func DisplayResults(results *BacktestResults) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().Println("BACKTEST RESULTS")
	pterm.Println()

	// Create table data
	data := pterm.TableData{
		{"Metric", "Value"},
		{"Total P&L", formatMoney(results.TotalPnL)},
		{"Win Rate", fmt.Sprintf("%.1f%%", results.WinRate)},
		{"Total Trades", fmt.Sprintf("%d", results.TotalTrades)},
		{"Avg Trade P&L", formatMoney(results.AvgTradePnL)},
		{"Sharpe Ratio", fmt.Sprintf("%.2f", results.SharpeRatio)},
		{"Max Drawdown", fmt.Sprintf("%.1f%%", results.MaxDrawdown)},
		{"Profit Factor", fmt.Sprintf("%.2f", results.ProfitFactor)},
		{"Duration", fmt.Sprintf("%.2fs", results.Duration.Seconds())},
	}

	pterm.DefaultTable.WithHasHeader().WithData(data).Render()

	if results.ResultsFile != "" {
		pterm.Println()
		Success(fmt.Sprintf("Results saved to: %s", pterm.Cyan(results.ResultsFile)))
	}
}

// DisplayConfigSummary shows a summary of the configuration
func DisplayConfigSummary(strategy, exchange, pair, timeframe string) {
	Success("Loading configuration")

	items := []pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Strategy: %s", pterm.Cyan(strategy))},
		{Level: 0, Text: fmt.Sprintf("Exchange: %s", pterm.Cyan(exchange))},
		{Level: 0, Text: fmt.Sprintf("Pair: %s", pterm.Cyan(pair))},
		{Level: 0, Text: fmt.Sprintf("Timeframe: %s", pterm.Cyan(timeframe))},
	}

	pterm.DefaultBulletList.WithItems(items).Render()
	pterm.Println()
}

// DisplayOverrides shows config overrides
func DisplayOverrides(overrides map[string]string) {
	if len(overrides) == 0 {
		return
	}

	Warning("Overriding config values:")
	for key, value := range overrides {
		pterm.Printf("  %s: %s\n", key, pterm.Cyan(value))
	}
	pterm.Println()
}

// CreateProgressBar creates a progress bar for backtest execution
func CreateProgressBar(description string, total int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "▕",
			BarEnd:        "▏",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetPredictTime(true),
	)
}

// DisplayDryRun shows what would run in dry-run mode
func DisplayDryRun(strategy, exchange, pair, timeframe string) {
	Info("Dry-run mode (no execution)")
	pterm.Println()

	pterm.DefaultBox.WithTitle("Configuration Preview").WithTitleTopCenter().Println(
		fmt.Sprintf("Strategy:      %s\n", pterm.Cyan(strategy)) +
			fmt.Sprintf("Exchange:      %s\n", pterm.Cyan(exchange)) +
			fmt.Sprintf("Pair:          %s\n", pterm.Cyan(pair)) +
			fmt.Sprintf("Timeframe:     %s\n", pterm.Cyan(timeframe)) +
			fmt.Sprintf("Execution:     %s", pterm.Cyan("Maximum speed (deterministic)")),
	)

	pterm.Println()
	Info("Estimated time: ~2 seconds")
	Info("Data size: ~150MB historical candles")
	pterm.Println()
	pterm.Println("Use --verbose for more details")
}

// DisplayError shows a formatted error with helpful hints
func DisplayError(title, reason string, fixes []string) {
	pterm.Println()
	pterm.Error.WithShowLineNumber(false).Println(title)
	pterm.Println()

	pterm.Printf("Reason: %s\n", pterm.Red(reason))
	pterm.Println()

	if len(fixes) > 0 {
		pterm.DefaultBasicText.Println("Fix this:")
		for i, fix := range fixes {
			pterm.Printf("  %d. %s\n", i+1, fix)
		}
		pterm.Println()
	}

	pterm.Printf("Documentation: %s\n", pterm.Blue("https://kronos.io/docs/config"))
}

// ShowNextSteps displays next steps after init
func ShowNextSteps(steps []string) {
	pterm.Println()
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.FgLightBlue)).Println("Next steps:")

	for i, step := range steps {
		pterm.Printf("  %d. %s\n", i+1, step)
	}
}

// formatMoney formats a float as money
func formatMoney(amount float64) string {
	sign := ""
	if amount > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s$%.2f", sign, amount)
}

// Confirm asks for user confirmation
func Confirm(message string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).Show(message)
	return result
}

// SelectFromList shows an interactive list selection
func SelectFromList(prompt string, options []string) (string, error) {
	selector := pterm.DefaultInteractiveSelect.WithOptions(options)
	selector.DefaultText = prompt
	return selector.Show()
}

// TextInput prompts for text input
func TextInput(prompt string, defaultValue string) string {
	result, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue(defaultValue).Show(prompt)
	return result
}
