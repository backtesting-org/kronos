package monitoring

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	"go.uber.org/fx"
)

// Module provides monitoring dependencies via FX
var Module = fx.Module("monitoring",
	fx.Provide(
		// ViewQuerier implementation - queries running instances via socket
		fx.Annotate(
			NewQuerier,
			fx.As(new(monitoring.ViewQuerier)),
		),
	),
)
