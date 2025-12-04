// Package monitoring provides types and interfaces for querying live strategy runtime data.
//
// This package contains:
//   - ViewRegistry interface (implemented by SDK to expose runtime data)
//   - ViewQuerier interface (implemented by CLI to query running strategy instances)
//   - StrategyMetrics type for runtime metrics
//
// Most types are passed through from the SDK directly (OrderBook, Trade, StrategyExecution, etc.)
package monitoring

import (
	"time"
)

// StrategyMetrics represents runtime metrics for a strategy
type StrategyMetrics struct {
	StrategyName     string        `json:"strategy_name"`
	Status           string        `json:"status"`
	LastSignalTime   time.Time     `json:"last_signal_time"`
	SignalsGenerated int           `json:"signals_generated"`
	SignalsExecuted  int           `json:"signals_executed"`
	SignalsFailed    int           `json:"signals_failed"`
	AverageLatency   time.Duration `json:"average_latency"`
	ActivePositions  int           `json:"active_positions"`
	DailyPnL         float64       `json:"daily_pnl"`
	WeeklyPnL        float64       `json:"weekly_pnl"`
	MonthlyPnL       float64       `json:"monthly_pnl"`
}
