package config

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"go.uber.org/fx"
)

var Module = fx.Module("config",
	fx.Provide(
		settings.NewConfiguration,
		strategy.NewStrategyConfigService,
	),
)
