package interactive

import (
	"fmt"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/pterm/pterm"
)

// InteractiveMode guides the user through configuration interactively
func InteractiveMode() (*config.Config, error) {
	ui.ShowBanner()
	ui.Success("Interactive Backtest Mode")
	pterm.Println()

	cfg := config.DefaultConfig()

	// Select Strategy
	strategy, err := ui.SelectFromList(
		"Select a strategy:",
		[]string{"market_making", "volume_maximizer", "arbitrage", "moving_average"},
	)
	if err != nil {
		return nil, err
	}
	cfg.Backtest.Strategy = strategy

	// Select Exchange
	exchange, err := ui.SelectFromList(
		"Select an exchange:",
		[]string{"binance", "coinbase", "kraken", "bybit"},
	)
	if err != nil {
		return nil, err
	}
	cfg.Backtest.Exchange = exchange

	// Select Pair
	pair, err := ui.SelectFromList(
		"Select a trading pair:",
		[]string{"BTC/USDT", "ETH/USDT", "SOL/USDT", "BNB/USDT"},
	)
	if err != nil {
		return nil, err
	}
	cfg.Backtest.Pair = pair

	// Select Date Range
	dateRange, err := ui.SelectFromList(
		"Select date range:",
		[]string{"Last 7 days", "Last 30 days", "Last 90 days", "Last 6 months", "Custom"},
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	switch dateRange {
	case "Last 7 days":
		cfg.Backtest.Timeframe.Start = now.AddDate(0, 0, -7).Format("2006-01-02")
		cfg.Backtest.Timeframe.End = now.Format("2006-01-02")
	case "Last 30 days":
		cfg.Backtest.Timeframe.Start = now.AddDate(0, 0, -30).Format("2006-01-02")
		cfg.Backtest.Timeframe.End = now.Format("2006-01-02")
	case "Last 90 days":
		cfg.Backtest.Timeframe.Start = now.AddDate(0, 0, -90).Format("2006-01-02")
		cfg.Backtest.Timeframe.End = now.Format("2006-01-02")
	case "Last 6 months":
		cfg.Backtest.Timeframe.Start = now.AddDate(0, -6, 0).Format("2006-01-02")
		cfg.Backtest.Timeframe.End = now.Format("2006-01-02")
	case "Custom":
		startDate := ui.TextInput("Start date (YYYY-MM-DD):", "2024-01-01")
		endDate := ui.TextInput("End date (YYYY-MM-DD):", now.Format("2006-01-02"))
		cfg.Backtest.Timeframe.Start = startDate
		cfg.Backtest.Timeframe.End = endDate
	}

	// Show preview
	pterm.Println()
	showConfigPreview(cfg)
	pterm.Println()

	// Confirm
	confirmed := ui.Confirm("Ready to run backtest?")
	if !confirmed {
		return nil, fmt.Errorf("backtest cancelled by user")
	}

	return cfg, nil
}

// showConfigPreview displays a preview of the configuration
func showConfigPreview(cfg *config.Config) {
	pterm.DefaultBox.WithTitle("Configuration Preview").WithTitleTopCenter().Println(
		fmt.Sprintf("Strategy:    %s\n", pterm.Cyan(cfg.Backtest.Strategy)) +
			fmt.Sprintf("Exchange:    %s\n", pterm.Cyan(cfg.Backtest.Exchange)) +
			fmt.Sprintf("Pair:        %s\n", pterm.Cyan(cfg.Backtest.Pair)) +
			fmt.Sprintf("Timeframe:   %s to %s",
				pterm.Cyan(cfg.Backtest.Timeframe.Start),
				pterm.Cyan(cfg.Backtest.Timeframe.End)),
	)
}
