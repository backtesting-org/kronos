package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/interactive"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	configPath      string
	outputFormat    string
	watch           bool
	dryRun          bool
	interactiveMode bool
	verbose         bool
)

var backtestCmd = &cobra.Command{
	Use:   "backtest",
	Short: "Run a backtest simulation",
	Long: `Run a backtest simulation of your trading strategy.
	
By default, opens interactive mode if no config is specified.
Use --non-interactive to force config file execution.

Examples:
  # Interactive mode (default if no config)
  kronos backtest
  
  # Run with config file
  kronos backtest --config kronos.yml
  
  # Non-interactive with config
  kronos backtest --non-interactive --config kronos.yml
  
  # Dry run (preview what would run)
  kronos backtest --dry-run
  
  # Watch mode (re-run on config changes)
  kronos backtest --watch`,
	RunE: runBacktest,
}

func init() {
	backtestCmd.Flags().StringVar(&configPath, "config", "kronos.yml", "Path to config file")
	backtestCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text|json")
	backtestCmd.Flags().BoolVar(&watch, "watch", false, "Watch config file and re-run on changes")
	backtestCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would run without executing")
	backtestCmd.Flags().BoolVar(&interactiveMode, "interactive", false, "Force interactive mode (guided setup)")
	backtestCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose logging")
}

func runBacktest(cmd *cobra.Command, args []string) error {
	if watch {
		return runWatchMode()
	}

	// If interactive flag is set OR (no config file exists AND not non-interactive mode)
	if interactiveMode || (!config.FileExists(configPath) && !nonInteractive) {
		return executeInteractiveBacktest()
	}

	return executeBacktest()
}

func executeInteractiveBacktest() error {
	cfg, err := interactive.InteractiveMode()
	if err != nil {
		return err
	}

	// Run backtest with interactive config
	results, err := runBacktestSimulation(cfg)
	if err != nil {
		ui.Error(fmt.Sprintf("Backtest failed: %v", err))
		return err
	}

	if outputFormat == "json" {
		return displayResultsJSON(results)
	}

	ui.DisplayResults(results)
	return nil
}

func executeBacktest() error {
	var cfg *config.Config
	var err error

	// Load config from file
	if !config.FileExists(configPath) {
		ui.DisplayError(
			"Config file not found",
			fmt.Sprintf("kronos.yml does not exist in current directory"),
			[]string{
				"Run: kronos init",
				"Or specify config path: kronos backtest --config path/to/config.yml",
			},
		)
		return fmt.Errorf("config file not found: %s", configPath)
	}

	cfg, err = config.LoadConfig(configPath)
	if err != nil {
		ui.DisplayError(
			"Failed to load config",
			err.Error(),
			[]string{
				"Check your YAML syntax",
				"See example: https://kronos.io/docs/config",
			},
		)
		return err
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		ui.DisplayError(
			"Invalid configuration",
			err.Error(),
			[]string{
				"Fix the configuration in kronos.yml",
				"Run: kronos backtest --dry-run to validate",
			},
		)
		return err
	}

	// Format timeframe
	timeframe := fmt.Sprintf("%s to %s", cfg.Backtest.Timeframe.Start, cfg.Backtest.Timeframe.End)

	// Dry run mode
	if dryRun {
		ui.DisplayDryRun(
			cfg.Backtest.Strategy,
			cfg.Backtest.Exchange,
			cfg.Backtest.Pair,
			timeframe,
		)
		return nil
	}

	// Display config summary
	if !interactiveMode {
		ui.DisplayConfigSummary(
			cfg.Backtest.Strategy,
			cfg.Backtest.Exchange,
			cfg.Backtest.Pair,
			timeframe,
		)
	}

	// Run backtest
	results, err := runBacktestSimulation(cfg)
	if err != nil {
		ui.Error(fmt.Sprintf("Backtest failed: %v", err))
		return err
	}

	// Display results
	if outputFormat == "json" {
		return displayResultsJSON(results)
	}

	ui.DisplayResults(results)
	return nil
}

func runBacktestSimulation(cfg *config.Config) (*ui.BacktestResults, error) {
	// Create progress bar
	bar := ui.CreateProgressBar("Running backtest", 100)

	// Simulate backtest execution
	// In real implementation, this would call the kronos-sdk
	startTime := time.Now()

	// Simulate progress (runs as fast as possible)
	for i := 0; i <= 100; i++ {
		bar.Add(1)
		// Small delay just for visual feedback in simulation
		time.Sleep(20 * time.Millisecond)
	}

	duration := time.Since(startTime)

	// Generate simulated results (in real impl, this comes from SDK)
	results := &ui.BacktestResults{
		TotalPnL:     2340.50 + rand.Float64()*1000 - 500,
		WinRate:      65.0 + rand.Float64()*10,
		TotalTrades:  40 + rand.Intn(20),
		AvgTradePnL:  49.80 + rand.Float64()*20 - 10,
		SharpeRatio:  1.2 + rand.Float64()*0.5,
		MaxDrawdown:  -(1.5 + rand.Float64()*2),
		ProfitFactor: 1.8 + rand.Float64()*0.8,
		Duration:     duration,
	}

	// Save results if configured
	if cfg.Backtest.Output.SaveResults {
		resultsFile, err := saveResults(cfg, results)
		if err != nil {
			ui.Warning(fmt.Sprintf("Failed to save results: %v", err))
		} else {
			results.ResultsFile = resultsFile
		}
	}

	return results, nil
}

func saveResults(cfg *config.Config, results *ui.BacktestResults) (string, error) {
	// Create results directory
	if err := os.MkdirAll(cfg.Backtest.Output.ResultsDir, 0755); err != nil {
		return "", err
	}

	// Generate filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.json", cfg.Backtest.Strategy, timestamp)
	filepath := filepath.Join(cfg.Backtest.Output.ResultsDir, filename)

	// Create results data
	data := map[string]interface{}{
		"config":    cfg,
		"results":   results,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Write to file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", err
	}

	return filepath, nil
}

func displayResultsJSON(results *ui.BacktestResults) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func runWatchMode() error {
	ui.Info(fmt.Sprintf("Watching %s for changes...", configPath))
	ui.Info("Press Ctrl+C to stop")
	fmt.Println()

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add config file to watcher
	if err := watcher.Add(configPath); err != nil {
		return err
	}

	// Run initial backtest
	if err := executeBacktest(); err != nil {
		ui.Error(fmt.Sprintf("Initial backtest failed: %v", err))
	}

	// Watch for changes
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println()
				ui.Info(fmt.Sprintf("Config changed (%s)", time.Now().Format("2006-01-02 15:04:05")))
				ui.Info("Re-running backtest...")
				fmt.Println()

				if err := executeBacktest(); err != nil {
					ui.Error(fmt.Sprintf("Backtest failed: %v", err))
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			ui.Error(fmt.Sprintf("Watcher error: %v", err))
		}
	}
}
