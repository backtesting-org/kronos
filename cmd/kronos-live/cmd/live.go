package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/live/runtime"
	"github.com/spf13/cobra"
)

type LiveCommand struct {
	Cmd     *cobra.Command
	runtime runtime.Runtime
}

func NewLiveCommand(rt runtime.Runtime) *LiveCommand {
	lc := &LiveCommand{
		runtime: rt,
	}

	lc.Cmd = &cobra.Command{
		Use:   "run",
		Short: "Run a live trading strategy",
		Long:  `Executes a compiled strategy plugin with configured exchange connectors from exchanges.yml`,
		RunE:  lc.run,
	}

	lc.Cmd.Flags().String("strategy-dir", "", "Path to strategy directory (required)")
	lc.Cmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no real trades)")

	_ = lc.Cmd.MarkFlagRequired("strategy-dir")

	return lc
}

func (lc *LiveCommand) run(cmd *cobra.Command, _ []string) error {
	strategyDir, _ := cmd.Flags().GetString("strategy-dir")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// TODO: Pass dry-run flag to runtime when SDK supports it

	// Check if strategy directory exists
	if _, err := os.Stat(strategyDir); os.IsNotExist(err) {
		return fmt.Errorf("strategy directory not found: %s", strategyDir)
	}

	// Setup signal handling
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
	fmt.Printf("   Strategy: %s\n", strategyDir)
	if dryRun {
		fmt.Printf("   Mode: DRY RUN (no real trades)\n")
	} else {
		fmt.Printf("   Mode: LIVE TRADING\n")
	}
	fmt.Println("\nPress Ctrl+C to stop...\n")

	if err := lc.runtime.Run(ctx, strategyDir); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	fmt.Println("\n✅ Strategy stopped successfully")
	return nil
}
