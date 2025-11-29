package strategies

import (
	"github.com/backtesting-org/kronos-cli/internal/strategies/browse"
	"github.com/backtesting-org/kronos-cli/internal/strategies/compile"
	"go.uber.org/fx"
)

var Module = fx.Options(
	browse.Module,
	compile.Module,
)
