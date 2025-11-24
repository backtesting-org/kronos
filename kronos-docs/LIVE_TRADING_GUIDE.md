# Kronos Live Trading

Deploy trading strategies to live exchanges using an interactive TUI with separated global and strategy-specific configurations.

## Configuration Structure

Kronos uses two levels of configuration:

1. **Global Exchange Config** (`exchanges.yml`) - Credentials and exchange settings shared across all strategies
2. **Strategy Config** (`strategies/*/config.yml`) - Strategy-specific settings (assets, parameters, risk)

## Quick Start

### 1. Create global exchanges config

Create `exchanges.yml` at your project root:

```yaml
exchanges:
  - name: paradex
    enabled: true
    network: mainnet  # or testnet
    credentials:
      account_address: ""
      eth_private_key: ""
      l2_private_key: ""

  - name: bybit
    enabled: true
    credentials:
      api_key: ""
      api_secret: ""
```

See [exchanges.yml.example](exchanges.yml.example) for more options.

### 2. Create strategy with config.yml

```bash
mkdir -p strategies/my-strategy
cd strategies/my-strategy
```

Create `config.yml`:

```yaml
name: my-strategy
description: My trading strategy
status: ready

# Reference exchanges from exchanges.yml
exchanges:
  - paradex

# Assets per exchange
assets:
  paradex:
    - BTC-USD-PERP
    - ETH-USD-PERP

# Strategy parameters
parameters:
  lookback_period: 20
  entry_threshold: 0.02

# Risk management
risk:
  max_position_size: 10000.0
  max_daily_loss: 1000.0

execution:
  dry_run: true
  mode: live
```

See [config.yml.example](config.yml.example) for full template.

### 3. Build your strategy plugin

```bash
go build -buildmode=plugin -o my-strategy.so .
```

### 4. Run live trading

```bash
cd ../..  # Back to project root
kronos live
```

## Interactive Flow

The TUI guides you through:

1. **Select Strategy** - Choose from strategies in `./strategies/`
2. **Select Exchange** - Pick which exchange to use (if multiple configured)
3. **Enter Credentials** - Input API keys (masked for security, pre-filled if already saved)
4. **Confirm** - Type "CONFIRM" to proceed
5. **Execute** - Press Enter to start live trading

## Configuration Benefits

### Why Two Config Files?

**Global (`exchanges.yml`)**:
- Store credentials once, use across all strategies
- One file to secure/backup
- Easy credential rotation
- Enable/disable exchanges globally

**Strategy-specific (`config.yml`)**:
- Different assets per strategy
- Strategy-specific parameters
- Independent risk management
- Easy to version control (no credentials)

### Example: Multiple Strategies, One Exchange

```
my-project/
├── exchanges.yml          # Paradex credentials here (once)
└── strategies/
    ├── momentum/
    │   └── config.yml     # Uses: paradex, BTC-USD-PERP
    ├── grid/
    │   └── config.yml     # Uses: paradex, ETH-USD-PERP
    └── arbitrage/
        └── config.yml     # Uses: paradex + bybit, multiple assets
```

All three strategies share the same Paradex credentials from `exchanges.yml`!

## Manual Editing

You can edit configs directly:

**Add credentials manually**:
```bash
# Edit exchanges.yml
vim exchanges.yml
# Add your API keys under credentials section
```

**Update strategy assets**:
```bash
# Edit strategy config
vim strategies/momentum/config.yml
# Modify assets list
```

**Or use the CLI**: Run `kronos live` and enter credentials interactively - they'll be saved to `exchanges.yml`.

## What it does

- Reads global exchange config from `./exchanges.yml`
- Discovers strategies from `./strategies/*/config.yml`
- Merges configs (strategy references + global credentials)
- Prompts for missing credentials
- Saves credentials to `exchanges.yml`
- Executes `kronos-live` with correct flags

Example command executed for Paradex:
```bash
kronos-live run \
  --exchange paradex \
  --strategy ./strategies/momentum/momentum.so \
  --paradex-account-address 0x123... \
  --paradex-eth-private-key 0xabc... \
  --paradex-network mainnet
```
