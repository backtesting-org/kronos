package main

import (
	"context"
	"log"

	"github.com/backtesting-org/kronos-cli/cmd"
	"github.com/backtesting-org/kronos-cli/internal/app"
	"go.uber.org/fx"
)

func main() {
	fxApp := fx.New(
		app.Module,
		cmd.Module, // Use the command module
		fx.Invoke(runCLI),
		//fx.NopLogger, // Suppress fx startup logs
	)

	fxApp.Run()
}

func runCLI(lc fx.Lifecycle, shutdowner fx.Shutdowner, root *cmd.RootCommand) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run the CLI in a goroutine so fx can manage lifecycle
			go func() {
				if err := root.Cmd.Execute(); err != nil {
					log.Printf("Error executing command: %v\n", err)
					log.Fatal(err)
				}
				// Shut down fx app after CLI command completes
				if err := shutdowner.Shutdown(); err != nil {
					log.Printf("Error shutting down: %v\n", err)
				}
			}()
			return nil
		},
	})
}
