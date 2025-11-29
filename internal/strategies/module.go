package strategies

import (
	"github.com/backtesting-org/kronos-cli/internal/strategies/browse"
	"go.uber.org/fx"
)

var Module = fx.Options(
	browse.Module,
)
