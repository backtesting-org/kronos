package monitoring

import (
	monitoring2 "github.com/backtesting-org/kronos-sdk/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	"go.uber.org/fx"
)

// Module provides monitoring dependencies via FX
var Module = fx.Module("monitoring",
	fx.Provide(
		// ViewRegistry implementation - pulls data from SDK components
		fx.Annotate(
			monitoring2.NewViewRegistry,
			fx.As(new(monitoring.ViewRegistry)),
		),
		// ViewQuerier implementation - queries running instances via socket
		fx.Annotate(
			NewQuerier,
			fx.As(new(monitoring.ViewQuerier)),
		),
	),
)
