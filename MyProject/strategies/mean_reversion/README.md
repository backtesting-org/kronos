# Mean Reversion Strategy Example

A simple Bollinger Bands mean reversion strategy demonstrating the Kronos SDK.

## What This Strategy Does

This is a **Mean Reversion** strategy that:
- Uses Bollinger Bands (20-period, 2 standard deviations) to identify extremes
- Uses RSI (14-period) for additional confirmation
- **Buys** when price touches the lower band AND RSI < 35 (oversold)
- **Sells** when price touches the upper band AND RSI > 65 (overbought)
- Targets the middle band (20-period SMA) for mean reversion

## Strategy Logic

```go
// Buy at lower band with RSI confirmation
if price.LessThan(bb.Lower) && rsi.LessThan(decimal.NewFromInt(35)) {
    return BUY signal
}

// Sell at upper band with RSI confirmation  
if price.GreaterThan(bb.Upper) && rsi.GreaterThan(decimal.NewFromInt(65)) {
    return SELL signal
}
```

## Key Features Demonstrated

1. **Multiple Indicators** - Combines Bollinger Bands with RSI
2. **Error Handling** - Properly handles indicator calculation errors
3. **Logging** - Uses structured logging for opportunities
4. **Signal Builder** - Uses the fluent signal builder API
5. **Type Safety** - Demonstrates decimal precision and type-safe assets

## Running the Strategy

### Backtest

```bash
cd examples/cash_carry
kronos backtest --strategy MeanReversion --start 2024-01-01 --end 2024-12-31
```

### Live Trading

```bash
kronos live --strategy MeanReversion --exchange binance
```

## Code Structure

```go
type MeanReversionStrategy struct {
    k *sdk.Kronos  // Kronos SDK instance
}

func (s *MeanReversionStrategy) GetSignals() ([]*strategy.Signal, error) {
    // 1. Get indicators
    bb := s.k.Indicators.BollingerBands(btc, 20, 2.0)
    price := s.k.Market.Price(btc)
    rsi := s.k.Indicators.RSI(btc, 14)
    
    // 2. Check conditions
    if price.LessThan(bb.Lower) && rsi.LessThan(35) {
        // 3. Create signal
        signal := s.k.Signal(s.GetName()).
            Buy(btc, connector.Binance, decimal.NewFromFloat(0.1)).
            Build()
        return []*strategy.Signal{signal}, nil
    }
    
    return nil, nil
}
```

## Strategy Characteristics

- **Type**: Mean Reversion
- **Risk Level**: Medium
- **Assets**: BTC (easily extended to multiple assets)
- **Indicators**: Bollinger Bands (20, 2.0), RSI (14)
- **Position Size**: 0.1 BTC per signal
- **Exchange**: Binance (configurable)

## Expected Performance

Mean reversion strategies work best in:
- **Ranging markets** - Price oscillates between support/resistance
- **High liquidity** - Tight spreads for quick execution
- **Mean-reverting assets** - Assets that return to average

May struggle in:
- **Strong trends** - Price breaks bands and keeps moving
- **Low liquidity** - Slippage reduces profitability
- **High volatility** - Stops get hit before reversion

## Improvements to Consider

1. **Dynamic Position Sizing** - Scale position based on how far price is from bands
2. **Trend Filter** - Only trade mean reversion in ranging markets
3. **Multiple Timeframes** - Confirm on higher timeframe
4. **Stop Loss** - Add ATR-based stops for risk management
5. **Multiple Assets** - Extend to trade BTC, ETH, SOL simultaneously
6. **Band Width Filter** - Avoid trades when bands are very narrow (low volatility)

## Learn More

- [Bollinger Bands Indicator](../../docs/api/indicators/bollinger-bands.md)
