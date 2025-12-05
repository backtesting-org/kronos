package strategies

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/browse"
	"github.com/backtesting-org/kronos-cli/internal/router"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
)

// StrategyBrowser handles browsing strategies and selecting actions
type StrategyBrowser interface {
}

type strategyBrowser struct {
	strategyService strategy.StrategyConfig
	compileService  strategyTypes.CompileService
	listFactory     browse.StrategyListViewFactory
	router          router.Router
}

func NewStrategyBrowser(
	strategyService strategy.StrategyConfig,
	compileService strategyTypes.CompileService,
	listFactory browse.StrategyListViewFactory,
	r router.Router,
) StrategyBrowser {
	return &strategyBrowser{
		strategyService: strategyService,
		compileService:  compileService,
		listFactory:     listFactory,
		router:          r,
	}
}
