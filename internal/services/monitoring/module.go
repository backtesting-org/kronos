package monitoring

import (
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"go.uber.org/fx"
)

// Module provides monitoring dependencies via FX
var Module = fx.Module("monitoring",
	fx.Provide(
		// ViewRegistry implementation - pulls data from SDK components
		fx.Annotate(
			NewViewRegistry,
			fx.As(new(monitoring.ViewRegistry)),
		),
	),
)
