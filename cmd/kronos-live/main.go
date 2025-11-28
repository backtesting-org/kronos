package main

import (
	"context"
	"log"

	"github.com/backtesting-org/kronos-cli/cmd/kronos-live/cmd"
	"github.com/backtesting-org/kronos-cli/internal/app"
	"github.com/backtesting-org/kronos-cli/internal/live/runtime"
	"go.uber.org/fx"
)

func main() {
	fxApp := fx.New(
		app.Module,
		runtime.Module,
		cmd.Module,
		fx.Invoke(runCLI),
		fx.NopLogger,
	)

	fxApp.Run()
}

func runCLI(lc fx.Lifecycle, shutdowner fx.Shutdowner, root *cmd.RootCommand) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := root.Cmd.Execute(); err != nil {
					log.Printf("Error executing command: %v\n", err)
					log.Fatal(err)
				}
				shutdowner.Shutdown()
			}()
			return nil
		},
	})
}
