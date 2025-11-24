package main

import (
	sdk "github.com/backtesting-org/kronos-sdk/pkg/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/shopspring/decimal"
)

// MeanReversionStrategy implements a Bollinger Bands mean reversion strategy
type meanReversionStrategy struct {
	strategy.BaseStrategy
	k *sdk.Kronos
}

// NewMeanReversion creates a new mean reversion strategy instance
func NewMeanReversion(k *sdk.Kronos) strategy.Strategy {
	return &meanReversionStrategy{k: k}
}

// GetSignals generates trading signals based on Bollinger Bands mean reversion
func (s *meanReversionStrategy) GetSignals() ([]*strategy.Signal, error) {
	btc := s.k.Asset("BTC")

	// Get indicators
	bb, err := s.k.Indicators.BollingerBands(btc, 20, 2.0)
	if err != nil {
		s.k.Log().Debug("MeanReversion", "BTC", "Failed to get Bollinger Bands: %v", err)
		return nil, nil
	}

	price, err := s.k.Market.Price(btc)
	if err != nil {
		s.k.Log().Debug("MeanReversion", "BTC", "Failed to get price: %v", err)
		return nil, nil
	}

	rsi, err := s.k.Indicators.RSI(btc, 14)
	if err != nil {
		s.k.Log().Debug("MeanReversion", "BTC", "Failed to get RSI: %v", err)
		return nil, nil
	}

	// Buy at lower band with RSI confirmation
	if price.LessThan(bb.Lower) && rsi.LessThan(decimal.NewFromInt(35)) {
		s.k.Log().Opportunity("MeanReversion", "BTC",
			"Price below lower band (%.2f), RSI oversold (%.2f), targeting middle band (%.2f)",
			price, rsi, bb.Middle)

		signal := s.k.Signal(s.GetName()).
			Buy(btc, connector.Binance, decimal.NewFromFloat(0.1)).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	// Sell at upper band with RSI confirmation
	if price.GreaterThan(bb.Upper) && rsi.GreaterThan(decimal.NewFromInt(65)) {
		s.k.Log().Opportunity("MeanReversion", "BTC",
			"Price above upper band (%.2f), RSI overbought (%.2f), targeting middle band (%.2f)",
			price, rsi, bb.Middle)

		signal := s.k.Signal(s.GetName()).
			Sell(btc, connector.Binance, decimal.NewFromFloat(0.1)).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	return nil, nil
}

// Interface implementation
func (s *meanReversionStrategy) GetName() strategy.StrategyName {
	return "Mean Reversion"
}

func (s *meanReversionStrategy) GetDescription() string {
	return "Bollinger Bands mean reversion with RSI confirmation"
}

func (s *meanReversionStrategy) GetRiskLevel() strategy.RiskLevel {
	return strategy.RiskLevelMedium
}

func (s *meanReversionStrategy) GetStrategyType() strategy.StrategyType {
	return strategy.StrategyTypeMeanReversion
}
