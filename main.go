package main

import (
	"context"
	"log"

	"github.com/backtesting-org/kronos-cli/cmd"
	"github.com/backtesting-org/kronos-cli/internal/app"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		app.Module,
		fx.Provide(cmd.NewCommands),
		fx.Invoke(runCLI),
	)

	app.Run()
}

func runCLI(lc fx.Lifecycle, commands *cmd.Commands) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run the CLI in a goroutine so fx can manage lifecycle
			go func() {
				if err := commands.Root.Execute(); err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
	})
}
