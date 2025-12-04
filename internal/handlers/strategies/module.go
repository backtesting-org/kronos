package strategies

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/browse"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/compile"
	"go.uber.org/fx"
)

var Module = fx.Options(
	browse.Module,
	compile.Module,
)
