package browse

import (
	"go.uber.org/fx"
)

// Module provides browse view constructors in DI
var Module = fx.Module("browse",
	fx.Provide(
		NewStrategyDetailView,
		NewStrategyListView,
	),
)
