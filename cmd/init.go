package cmd

import (
	"fmt"
	"os"

	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kronos.yml configuration file",
	Long: `Create a kronos.yml configuration file with sensible defaults.
	
This creates a template configuration that you can customize for your backtests.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Show banner
	ui.ShowBanner()

	// Check if kronos.yml already exists
	if config.FileExists("kronos.yml") {
		ui.Warning("kronos.yml already exists in current directory")
		overwrite := ui.Confirm("Overwrite existing file?")
		if !overwrite {
			ui.Info("Init cancelled")
			return nil
		}
	}

	// Create default config
	cfg := config.DefaultConfig()

	// Write to YAML
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("kronos.yml", yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Success message
	pterm.Println()
	ui.Success("Kronos CLI initialized")
	ui.Success("Created kronos.yml with default configuration")
	pterm.Println()

	// Show config preview
	pterm.DefaultBox.WithTitle("Sample kronos.yml created").WithTitleTopCenter().Println(
		fmt.Sprintf("strategy:    %s\n", pterm.Cyan(cfg.Backtest.Strategy)) +
			fmt.Sprintf("exchange:    %s\n", pterm.Cyan(cfg.Backtest.Exchange)) +
			fmt.Sprintf("pair:        %s\n", pterm.Cyan(cfg.Backtest.Pair)) +
			fmt.Sprintf("timeframe:   %s to %s",
				pterm.Cyan(cfg.Backtest.Timeframe.Start),
				pterm.Cyan(cfg.Backtest.Timeframe.End)),
	)

	// Next steps
	ui.ShowNextSteps([]string{
		"Edit kronos.yml to customize your configuration",
		"Run: " + pterm.Cyan("kronos backtest"),
		"Or try interactive mode: " + pterm.Cyan("kronos backtest --interactive"),
		"View results: " + pterm.Cyan("kronos analyze"),
	})

	return nil
}
