# MyProject

A Kronos trading strategy project using the mean_reversion strategy.

## Setup

1. Configure your exchange credentials in `exchanges.yml`
2. Install dependencies: `go mod tidy`
3. Run the strategy: `go run strategies/mean_reversion/strategy.go`

## Configuration

- `exchanges.yml` - Global exchange and asset configuration
- `strategies/mean_reversion/config.yml` - Strategy-specific parameters

## Documentation

For more information, visit: https://github.com/backtesting-org/kronos-sdk
