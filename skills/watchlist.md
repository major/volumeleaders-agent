# watchlist

Manage saved watchlist configurations.

## watchlist configs

List all saved watchlist configurations. No flags.

```bash
volumeleaders-agent watchlist configs
```

Output fields: `SearchTemplateKey` (use as `--key` or `--watchlist-key` for other commands), `Name`, `Tickers`, plus filter settings (volume/price/dollar ranges, RSI conditions, trade type booleans, `SecurityTypeKey`, `MinVCD`, `SectorIndustry`).

## watchlist tickers

Get tickers and summary data from a watchlist.

Optional: `--watchlist-key` (default -1 = all watchlists). Get key from `watchlist configs` `SearchTemplateKey`.

```bash
volumeleaders-agent watchlist tickers --watchlist-key 12345
```

Output fields: `Ticker`, `Price`, `NearestTop10TradeDate`, `NearestTop10TradeClusterDate`, `NearestTop10TradeLevel`

## watchlist delete

Required: `--key` (SearchTemplateKey from `watchlist configs`)

```bash
volumeleaders-agent watchlist delete --key 12345
```

## watchlist add-ticker

Add a single ticker to an existing watchlist.

Required: `--watchlist-key`, `--ticker`

```bash
volumeleaders-agent watchlist add-ticker --watchlist-key 12345 --ticker AAPL
```

## watchlist create

Required: `--name`

Optional flags:

| Category | Flags | Defaults |
|---|---|---|
| **Tickers** | `--tickers` | (none, comma-separated, max 500) |
| **Volume** | `--min-volume`, `--max-volume` | 0, 2000000000 |
| **Dollars** | `--min-dollars`, `--max-dollars` | 0, 30000000000 |
| **Price** | `--min-price`, `--max-price` | 0, 100000 |
| **Quality** | `--min-vcd`, `--min-relative-size`, `--max-trade-rank` | 0, 0 (0/5/10/25/50/100), -1 (-1/1/3/5/10/25/50/100) |
| **Security** | `--security-type`, `--sector-industry` | -1 (-1=all/1=stocks/26=ETFs/4=REITs), (none, max 100 chars) |
| **Trade types** | `--normal-prints`, `--signature-prints`, `--late-prints`, `--timely-prints`, `--dark-pools`, `--lit-exchanges`, `--sweeps`, `--blocks` | all true. Disable with `--no-{name}` |
| **Sessions** | `--premarket-trades`, `--rth-trades`, `--ah-trades`, `--opening-trades`, `--closing-trades`, `--phantom-trades`, `--offsetting-trades` | all true. Disable with `--no-{name}` |
| **RSI** | `--rsi-overbought-daily`, `--rsi-overbought-hourly`, `--rsi-oversold-daily`, `--rsi-oversold-hourly` | all -1 (1=yes, 0=no, -1=ignore) |

```bash
volumeleaders-agent watchlist create --name "Tech Sweeps" --tickers "AAPL,MSFT,GOOGL" --no-dark-pools --min-dollars 1000000
```

## watchlist edit

Required: `--key` (SearchTemplateKey)
All flags from `watchlist create` available. `--name` is optional for edit. Unspecified flags reset to their defaults.

```bash
volumeleaders-agent watchlist edit --key 12345 --name "Updated Watchlist"
```
