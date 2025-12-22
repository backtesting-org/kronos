package monitoring

import (
	"context"

	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/profiling"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

type viewRegistry struct {
	kronos           kronos.Kronos
	health           health.HealthStore
	strategyRegistry registry.StrategyRegistry
	profilingStore   profiling.ProfilingStore
}

func NewViewRegistry(
	health health.HealthStore,
	kronos kronos.Kronos,
	strategyRegistry registry.StrategyRegistry,
	profilingStore profiling.ProfilingStore,
) monitoring.ViewRegistry {
	return &viewRegistry{
		health:           health,
		kronos:           kronos,
		strategyRegistry: strategyRegistry,
		profilingStore:   profilingStore,
	}
}

// getStrategyName returns the single registered strategy name
func (r *viewRegistry) getStrategyName() strategy.StrategyName {
	strategies := r.strategyRegistry.GetAllStrategies()
	if len(strategies) == 0 {
		return ""
	}
	return strategies[0].GetName()
}

func (r *viewRegistry) GetPnLView() *monitoring.PnLView {
	ctx := context.Background()
	name := r.getStrategyName()
	if name == "" {
		return nil
	}

	pnl := r.kronos.Activity().PNL()
	realizedPnL := pnl.GetRealizedPNL(ctx, name)
	unrealizedPnL, _ := pnl.GetUnrealizedPNL(ctx, name)
	totalPnL, _ := pnl.GetTotalPNL(ctx)
	totalFees := pnl.GetFeesByStrategy(ctx, name)

	return &monitoring.PnLView{
		StrategyName:  string(name),
		RealizedPnL:   realizedPnL,
		UnrealizedPnL: unrealizedPnL,
		TotalPnL:      totalPnL,
		TotalFees:     totalFees,
	}
}

func (r *viewRegistry) GetPositionsView() *strategy.StrategyExecution {
	ctx := context.Background()
	name := r.getStrategyName()
	if name == "" {
		return nil
	}
	return r.kronos.Activity().Positions().GetStrategyExecution(ctx, name)
}

func (r *viewRegistry) GetOrderbookView(symbol string) *connector.OrderBook {
	ctx := context.Background()

	asset := r.kronos.Asset(symbol)
	ob, err := r.kronos.Market().OrderBook(ctx, asset)
	if err != nil {
		return nil
	}
	return ob
}

func (r *viewRegistry) GetRecentTrades(limit int) []connector.Trade {
	ctx := context.Background()

	name := r.getStrategyName()
	if name == "" {
		return nil
	}
	trades := r.kronos.Activity().Positions().GetTradesForStrategy(ctx, name)
	if len(trades) <= limit {
		return trades
	}
	return trades[len(trades)-limit:]
}

func (r *viewRegistry) GetMetrics() *monitoring.StrategyMetrics {
	name := r.getStrategyName()
	return &monitoring.StrategyMetrics{
		StrategyName: string(name),
		Status:       "running",
	}
}

func (r *viewRegistry) GetHealth() *health.SystemHealthReport {
	return r.health.GetSystemHealth()
}

func (r *viewRegistry) GetAvailableAssets() []monitoring.AssetExchange {
	universe := r.kronos.Universe()
	assets := universe.Assets
	exchanges := universe.Exchanges

	var result []monitoring.AssetExchange
	for asset, _ := range assets {
		for _, exchange := range exchanges {
			result = append(result, monitoring.AssetExchange{
				Asset:    asset.Symbol(),
				Exchange: string(exchange.Name),
			})
		}
	}
	return result
}

func (r *viewRegistry) GetProfilingStats() *monitoring.ProfilingStats {
	if r.profilingStore == nil {
		return nil
	}

	name := r.getStrategyName()
	if name == "" {
		return nil
	}

	stats := r.profilingStore.GetStats(string(name))
	return &stats
}

func (r *viewRegistry) GetRecentExecutions(limit int) []monitoring.ProfilingMetrics {
	if r.profilingStore == nil {
		return nil
	}

	name := r.getStrategyName()
	if name == "" {
		return nil
	}

	return r.profilingStore.GetRecentMetrics(string(name), limit)
}
