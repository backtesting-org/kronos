package config

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"go.uber.org/fx"
)

var Module = fx.Module("config",
	fx.Provide(
		settings.NewConfiguration,
	),
)
