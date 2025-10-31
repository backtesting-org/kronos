# Kronos CLI

Fast, deterministic backtesting and live trading for algorithmic strategies.

Install
```bash
go install github.com/backtesting-org/kronos-cli@latest
```

Or use Homebrew:

```bash
brew install kronos
```

Get Started
```bash

Initialize a new project
kronos init

Run a backtest
kronos backtest

Deploy to live trading
kronos live
```

Commands
| Command | Purpose |
| --- | --- |
| kronos init | Initialize new project with config |
| kronos backtest | Run backtest simulation |
| kronos live | Start live trading |
| kronos analyze | View backtest results |
| kronos version | Show version info |

Configuration
Edit kronos.yml to configure your strategy:

Example kronos.yml:
```yaml
version: "1.0"

backtest:
  strategy: market_making
  exchange: binance
  pair: BTC/USDT
  timeframe:
    start: 2024-01-01
    end: 2024-06-30

  parameters:
    bid_spread: 0.1
    ask_spread: 0.1
    order_size: 1.0

live:
  enabled: false
  exchange: binance
```

Flags
```bash
kronos backtest [FLAGS]

FLAGS:
--config string Config file (default: kronos.yml)
--speed int Simulation speed (default: from config)
--seed int64 Random seed for reproducibility
--output string Output format: text|json
--verbose Verbose logging
```

Output
```
✓ Loading kronos.yml
✓ Running backtest...

████████████████████ 100% [2.3s]

Results:
Total P&L: +$2,340.50
Win Rate: 68.5%
Trades: 47
Sharpe Ratio: 1.42
Max Drawdown: -2.3%

✓ Saved to: results/backtest_2025-10-31.json
```

Next Steps
📖 Read the docs

🔗 Join our Discord

🐛 Report issues

License
MIT
