package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/services/live/runtime"
	"github.com/spf13/cobra"
)

type RunStrategyCommand struct {
	Cmd     *cobra.Command
	runtime runtime.Runtime
}

func NewRunStrategyCommand(rt runtime.Runtime) *RunStrategyCommand {
	rsc := &RunStrategyCommand{
		runtime: rt,
	}

	rsc.Cmd = &cobra.Command{
		Use:    "run-strategy",
		Hidden: true, // Hidden from help - internal use only
		Short:  "Run a live trading strategy instance (internal command)",
		Long:   `Internal command used to run a strategy instance in a separate process.`,
		RunE:   rsc.run,
	}

	rsc.Cmd.Flags().String("strategy", "", "Strategy name (required)")
	_ = rsc.Cmd.MarkFlagRequired("strategy")

	return rsc
}

func (rsc *RunStrategyCommand) run(cmd *cobra.Command, _ []string) error {
	strategyName, _ := cmd.Flags().GetString("strategy")

	// Build strategy directory path using convention: ./strategies/{strategy-name}
	strategyDir := filepath.Join("strategies", strategyName)

	// Check if strategy directory exists
	if _, err := os.Stat(strategyDir); os.IsNotExist(err) {
		return fmt.Errorf("strategy directory not found: %s", strategyDir)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Received shutdown signal, stopping strategy...")
		cancel()
	}()

	// Start runtime - it will load config.yml from strategy dir and exchanges.yml from project root
	fmt.Printf("🚀 Starting live trading\n")
	fmt.Printf("   Strategy: %s\n", strategyName)
	fmt.Printf("   Path: %s\n", strategyDir)
	fmt.Println("\nPress Ctrl+C to stop...")

	if err := rsc.runtime.Run(ctx, strategyDir); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	fmt.Println("\n✅ Strategy stopped successfully")
	return nil
}
