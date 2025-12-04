package monitoring

import (
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

type viewRegistry struct {
	kronos           kronos.Kronos
	health           health.HealthStore
	strategyRegistry registry.StrategyRegistry
}

func NewViewRegistry(
	health health.HealthStore,
	k kronos.Kronos,
	strategyRegistry registry.StrategyRegistry,
) monitoring.ViewRegistry {
	return &viewRegistry{
		health:           health,
		kronos:           k,
		strategyRegistry: strategyRegistry,
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
	name := r.getStrategyName()
	if name == "" {
		return nil
	}

	pnl := r.kronos.Activity().PNL()
	realizedPnL := pnl.GetRealizedPNL(name)
	unrealizedPnL, _ := pnl.GetUnrealizedPNL(name)
	totalPnL, _ := pnl.GetTotalPNL()
	totalFees := pnl.GetFeesByStrategy(name)

	return &monitoring.PnLView{
		StrategyName:  string(name),
		RealizedPnL:   realizedPnL,
		UnrealizedPnL: unrealizedPnL,
		TotalPnL:      totalPnL,
		TotalFees:     totalFees,
	}
}

func (r *viewRegistry) GetPositionsView() *strategy.StrategyExecution {
	name := r.getStrategyName()
	if name == "" {
		return nil
	}
	return r.kronos.Activity().Positions().GetStrategyExecution(name)
}

func (r *viewRegistry) GetOrderbookView(symbol string) *connector.OrderBook {
	asset := r.kronos.Asset(symbol)
	ob, err := r.kronos.Market().OrderBook(asset)
	if err != nil {
		return nil
	}
	return ob
}

func (r *viewRegistry) GetRecentTrades(limit int) []connector.Trade {
	name := r.getStrategyName()
	if name == "" {
		return nil
	}
	trades := r.kronos.Activity().Positions().GetTradesForStrategy(name)
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
