package main

import (
	"context"
	"log"

	"github.com/backtesting-org/kronos-cli/cmd"
	"github.com/backtesting-org/kronos-cli/internal/app"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	// Create a no-op logger to suppress fx output
	logger := zap.NewNop()

	app := fx.New(
		app.Module,
		fx.Provide(cmd.NewCommands),
		fx.Invoke(runCLI),
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
	)

	app.Run()
}

func runCLI(lc fx.Lifecycle, shutdowner fx.Shutdowner, commands *cmd.Commands) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run the CLI in a goroutine so fx can manage lifecycle
			go func() {
				if err := commands.Root.Execute(); err != nil {
					log.Fatal(err)
				}
				// Shut down fx app after CLI command completes
				shutdowner.Shutdown()
			}()
			return nil
		},
	})
}
