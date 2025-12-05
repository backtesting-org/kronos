# Kronos CLI

<div align="center">

**Fast, deterministic backtesting and live trading for algorithmic strategies.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-online-blue)](https://documentation-chi-ecru.vercel.app/docs/intro)

*Build, backtest, and deploy trading strategies with a simple TUI interface.*

[Quick Start](https://documentation-chi-ecru.vercel.app/docs/getting-started/) • [Features](#features) • [Documentation](https://documentation-chi-ecru.vercel.app/docs/intro) • [Examples](https://documentation-chi-ecru.vercel.app/docs/examples/)

</div>

---

## 🚀 What is Kronos?

Kronos is a **low-code algorithmic trading framework** that lets you write strategies in Go and deploy them to live markets with confidence. Built on a plugin architecture with hot-reload support, Kronos enables rapid strategy development and deployment.

### What Kronos Does For You

✅ **Plugin-Based Strategy System** - Write strategies as Go plugins, compile once, deploy anywhere  
✅ **Interactive TUI** - Beautiful terminal interface for managing strategies and monitoring live trades  
✅ **Multi-Exchange Support** - Unified API across multiple exchanges (Hyperliquid perps stable, more coming soon)  
✅ **Real-Time Monitoring** - Live orderbook, PnL, positions, and trade data via Unix sockets  
✅ **Graceful Lifecycle Management** - HTTP-based process control for reliable starts and stops  
✅ **Production-Ready** - Deploy strategies to live markets with confidence  

---

## 📦 Installation

### Install via Go

```bash
go install github.com/backtesting-org/kronos-cli@latest
```

### Build from Source

```bash
git clone https://github.com/backtesting-org/kronos-cli
cd kronos-cli
go build -o kronos
sudo mv kronos /usr/local/bin/
```

### Verify Installation

```bash
kronos version
```

---

## ⚡ Quick Start

### 1. Initialize a New Project

```bash
mkdir my-trading-bot && cd my-trading-bot
kronos
# Navigate to: 🆕 Create New Project
```

This creates:
```
my-trading-bot/
├── config.yml              # Strategy configuration
├── exchanges.yml           # Exchange credentials & settings
└── strategies/
    └── momentum/
        ├── config.yml      # Strategy metadata
        └── main.go         # Strategy implementation
```

### 2. Write Your Strategy

Kronos strategies implement a simple interface:

```go
package main

import (
    "github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
    "github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

func NewStrategy(k kronos.Kronos) strategy.Strategy {
    return &myStrategy{k: k}
}

type myStrategy struct {
    strategy.BaseStrategy
    k kronos.Kronos
}

func (s *myStrategy) GetSignals() ([]*strategy.Signal, error) {
    // Your trading logic here
    price, _ := s.k.Market().Price(s.k.Asset("BTC"))
    rsi, _ := s.k.Indicators().RSI(s.k.Asset("BTC"), 14)
    
    if rsi.LessThan(numerical.NewFromInt(30)) {
        signal := s.k.Signal(s.GetName()).
            Buy(s.k.Asset("BTC"), "hyperliquid", numerical.NewFromFloat(0.1)).
            Build()
        return []*strategy.Signal{signal}, nil
    }
    
    return nil, nil
}

func (s *myStrategy) GetName() strategy.StrategyName {
    return "MyStrategy"
}
```

### 3. Configure Your Strategy

Edit `strategies/momentum/config.yml`:

```yaml
name: momentum
display_name: "Momentum Strategy"
description: "RSI-based momentum trading"
type: momentum

exchanges:
  - hyperliquid

assets:
  hyperliquid:
    - BTC/USDT
    - ETH/USDT

indicators:
  rsi:
    period: 14
    oversold: 30
    overbought: 70

parameters:
  position_size: 0.1
```

### 4. Compile Your Strategy

```bash
kronos
# Navigate to: Strategies → momentum → Compile
```

Kronos compiles your strategy into a `.so` plugin file with progress tracking.

### 5. Deploy to Live Trading

```bash
kronos
# Navigate to: Strategies → momentum → Start Live
```

Your strategy runs as a detached process, continuing even after you close the CLI.

### 6. Monitor Live Strategies

```bash
kronos
# Navigate to: Monitor
```

Real-time monitoring dashboard shows:
- **Overview**: Strategy status, uptime, health
- **Positions**: Active positions across exchanges
- **Orderbook**: Live orderbook depth
- **Trades**: Recent trade history
- **PnL**: Realized/unrealized profit & loss

### 7. Stop a Running Strategy

```bash
# From Monitor view:
# 1. Select running instance
# 2. Press [S]
# 3. Confirm "Yes, Stop"
```

Graceful HTTP-based shutdown ensures clean process termination.

---

## 🎯 Features

### Strategy Development

- **Plugin Architecture** - Strategies compile to Go plugins (.so files)
- **Hot Reload** - Update strategies without restarting the framework
- **Type-Safe API** - Full IDE support with autocomplete
- **Rich Indicators** - RSI, MACD, Bollinger Bands, EMA, SMA, and more
- **Multi-Asset** - Trade multiple assets simultaneously
- **Multi-Exchange** - Execute across multiple exchanges in one strategy


### Live Trading

- **Process Isolation** - Each strategy runs in its own process
- **Detached Execution** - Strategies continue after CLI closes
- **State Persistence** - Instance state survives CLI restarts
- **Real-Time Data** - WebSocket + REST hybrid ingestion
- **Position Tracking** - Automatic position reconciliation
- **Trade Backfill** - Recovers trades on restart

### Monitoring

- **Unix Socket Communication** - Fast, local IPC
- **HTTP API** - RESTful access to strategy data
- **Live Orderbook** - Real-time order book updates
- **PnL Tracking** - Realized and unrealized profit/loss
- **Health Checks** - System health and error reporting
- **Multi-Instance** - Monitor multiple strategies at once

### Exchange Support

| Exchange | Spot | Perpetual | Status |
|----------|------|-----------|--------|
| Hyperliquid | 🚧 | ✅ | Perps Stable |
| Bybit | 🚧 | 🚧 | In Development |
| Paradex | 🚧 | 🚧 | In Development |

### User Interface

- **Beautiful TUI** - Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Keyboard Navigation** - Vim-style keybindings (hjkl)
- **Responsive Design** - Adapts to terminal size
- **Color Coding** - Visual status indicators
- **Progress Tracking** - Real-time compilation and backtest progress

---

## 📚 Documentation

**Full Documentation**: [https://documentation-chi-ecru.vercel.app/docs/intro](https://documentation-chi-ecru.vercel.app/docs/intro)

### Key Resources

- [Introduction](https://documentation-chi-ecru.vercel.app/docs/intro#what-kronos-does-for-you) - Architecture overview
- [Strategy Development](https://documentation-chi-ecru.vercel.app/docs/strategies) - Writing strategies
- [SDK Reference](https://documentation-chi-ecru.vercel.app/docs/sdk) - API documentation
- [Exchange Configuration](https://documentation-chi-ecru.vercel.app/docs/exchanges) - Setting up exchanges
- [Live Trading](https://documentation-chi-ecru.vercel.app/docs/live) - Deployment guide

---

## 🎨 Screenshots

### Main Menu
```
╭─────────────────────────────────────────────╮
│            KRONOS CLI v0.1.0                │
│                                             │
│  What would you like to do?                 │
│                                             │
│  ▶ 📂 Strategies                            │
│    📊 Monitor                               │
│    ⚙️  Settings                             │
│    ℹ️  Help                                  │
│    🆕 Create New Project                    │
│                                             │
│  ↑↓/jk Navigate  ↵ Select  q Quit           │
╰─────────────────────────────────────────────╯
```

### Live Monitoring
```
╭─────────────────────────────────────────────────────────────────────╮
│ MONITOR                                                             │
│                                                                     │
│  STATUS  STRATEGY           PID     UPTIME    PNL        HEALTH     │
│  ────────────────────────────────────────────────────────────────   │
│  ✓ RUN   momentum           86697   2h 30m    +$125.50  ███████     │
│    STP   arbitrage          -       -         -$43.20   ─────       │
│                                                                     │
│ [↑↓] Navigate • [Enter] Details • [S] Stop • [R] Refresh • [Q] Back │
╰─────────────────────────────────────────────────────────────────────╯
```

### Orderbook View
```
╭─────────────────────────────────────────────╮
│ ORDERBOOK - BTC/USDT (hyperliquid)         │
│                                             │
│ ASKS                                        │
│ 43,251.50  ████████░░  0.5420  $23,456    │
│ 43,250.00  ██████░░░░  0.3210  $13,888    │
│ 43,249.50  ████░░░░░░  0.1890  $8,174     │
│                                             │
│ BIDS                                        │
│ 43,248.00  ██████████  0.8920  $38,577    │
│ 43,247.50  ████████░░  0.6540  $28,284    │
│ 43,247.00  ██████░░░░  0.4320  $18,683    │
│                                             │
│ Spread: $3.50 (0.008%)  Last: 43,248.75   │
╰─────────────────────────────────────────────╯
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      KRONOS CLI (TUI)                       │
│  • Strategy Browser  • Compiler  • Monitor  • Live Trading │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              INSTANCE MANAGER (Process Control)             │
│  • Start/Stop Strategies  • State Persistence               │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
┌──────────────────┐    ┌──────────────────┐
│  Strategy Process│    │  Strategy Process│
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │ Your Plugin│ │    │  │ Your Plugin│ │
│  └────────────┘ │    │  └────────────┘ │
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │  SDK Core  │ │    │  │  SDK Core  │ │
│  └────────────┘ │    │  └────────────┘ │
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │ Monitoring │ │    │  │ Monitoring │ │
│  │   Server   │ │    │  │   Server   │ │
│  └────────────┘ │    │  └────────────┘ │
└────────┬─────────┘    └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     ▼
         ┌─────────────────────┐
         │   EXCHANGES         │
         │  • Hyperliquid      │
         │  • Bybit            │
         │  • Paradex          │
         └─────────────────────┘
```

### Key Components

1. **CLI** - Interactive terminal interface for managing strategies
2. **Instance Manager** - Controls strategy lifecycle (start/stop/monitor)
3. **Strategy Plugins** - Your compiled trading logic (.so files)
4. **SDK Runtime** - Core execution engine with data ingestion
5. **Monitoring Server** - HTTP API exposed via Unix sockets
6. **Exchange Connectors** - Unified interface to multiple exchanges

---

## 🔧 Commands

### Interactive Mode (Default)

```bash
kronos
```

Launches the TUI interface with full navigation.

### CLI Mode

```bash
kronos --cli          # Show command help
kronos version        # Show version info
```

### Advanced Usage

```bash
# Run specific strategy (internal use - called by Instance Manager)
kronos run-strategy --strategy momentum
```

---

## 📊 Example Strategies

### Momentum (RSI-Based)

```go
func (s *momentumStrategy) GetSignals() ([]*strategy.Signal, error) {
    asset := s.k.Asset("BTC")
    rsi, _ := s.k.Indicators().RSI(asset, 14)
    
    if rsi.LessThan(numerical.NewFromInt(30)) {
        return []*strategy.Signal{
            s.k.Signal(s.GetName()).
                Buy(asset, "hyperliquid", numerical.NewFromFloat(0.1)).
                Build(),
        }, nil
    }
    
    if rsi.GreaterThan(numerical.NewFromInt(70)) {
        return []*strategy.Signal{
            s.k.Signal(s.GetName()).
                Sell(asset, "hyperliquid", numerical.NewFromFloat(0.1)).
                Build(),
        }, nil
    }
    
    return nil, nil
}
```

### Arbitrage (Cross-Exchange)

```go
func (s *arbitrageStrategy) GetSignals() ([]*strategy.Signal, error) {
    asset := s.k.Asset("BTC")
    
    // Find arbitrage opportunities
    opportunities := s.k.Market().FindArbitrage(
        asset,
        numerical.NewFromFloat(0.5), // Min 0.5% spread
    )
    
    if len(opportunities) > 0 {
        opp := opportunities[0]
        return []*strategy.Signal{
            s.k.Signal(s.GetName()).
                Buy(asset, opp.BuyExchange, quantity).
                Sell(asset, opp.SellExchange, quantity).
                Build(),
        }, nil
    }
    
    return nil, nil
}
```

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Development Setup

```bash
# Clone the repo
git clone https://github.com/backtesting-org/kronos-cli
cd kronos-cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o kronos
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/services/monitoring/...

# With coverage
go test -cover ./...

# Watch mode
ginkgo watch -r
```

### Code Style

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write tests with [Ginkgo](https://github.com/onsi/ginkgo) and [Gomega](https://github.com/onsi/gomega)
- Use [mockery](https://github.com/vektra/mockery) for mocks

### Areas We Need Help

- 🌐 Additional exchange connectors
- 📊 More technical indicators
- 📚 Documentation improvements
- 🐛 Bug fixes and testing
- 🎨 UI/UX enhancements

---

## 🗺️ Roadmap

### v0.2.0 (Q1 2026)

- [ ] WebSocket-based live monitoring dashboard
- [ ] Strategy performance comparison tool
- [ ] Paper trading mode
- [ ] Discord/Telegram notifications
- [ ] Portfolio optimization tools

### v0.3.0 (Q2 2026)

- [ ] Cloud deployment support
- [ ] Strategy marketplace
- [ ] Multi-user support
- [ ] Advanced risk management
- [ ] Machine learning integration

### v1.0.0 (Q3 2026)

- [ ] Enterprise features
- [ ] Professional support
- [ ] Advanced analytics suite

---

## ❓ FAQ

**Q: Is Kronos suitable for production trading?**  
A: Yes, but use appropriate risk management. Start with small positions and paper trading.

**Q: What exchanges are supported?**  
A: Currently only Hyperliquid perpetuals are stable in production. Bybit and Paradex are in active development.

**Q: Can I run multiple strategies simultaneously?**  
A: Yes! Each strategy runs in its own isolated process.

**Q: How do I handle API keys securely?**  
A: Store them in `exchanges.yml` with proper file permissions (chmod 600).

**Q: Can I write strategies in languages other than Go?**  
A: Kronos strategies must be written in Go to compile as plugins. However, you can integrate machine learning models from any language using:
- **gRPC** - Call ML inference services in Python, R, or any language
- **ONNX Runtime** - Load pre-trained models directly in Go
- **HTTP APIs** - Connect to external prediction services
---

## 🐛 Troubleshooting

### Strategy Won't Compile

```bash
# Ensure Go version matches
go version  # Should be 1.24+

# Clear build cache
go clean -cache

# Rebuild with verbose output
go build -v -o strategies/momentum/momentum.so -buildmode=plugin strategies/momentum/main.go
```

### Process Won't Stop

```bash
# Find the process
ps aux | grep momentum

# Force kill
kill -9 <PID>

# Clean up socket files
rm ~/.kronos/sockets/*.sock
```

### "Already Running" Error

```bash
# Check for orphaned processes
ps aux | grep kronos

# Clean up state files
rm ~/.kronos/instances/*/state.json
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Cobra](https://github.com/spf13/cobra) - CLI commands
- [Fx](https://github.com/uber-go/fx) - Dependency injection
- [Ginkgo](https://github.com/onsi/ginkgo) - Testing framework

---

## 📞 Support

- 📖 [Documentation](https://documentation-chi-ecru.vercel.app/docs/intro)
- 💬 [Discord Community](#) *(coming soon)*
- 🐛 [Issue Tracker](https://github.com/backtesting-org/kronos-cli/issues)
- ✉️ [Email Support](#) *(for enterprise)*

---

<div align="center">

**⭐ If you find Kronos useful, please consider starring the repo! ⭐**

Made with ❤️ by the Kronos Team

</div>
