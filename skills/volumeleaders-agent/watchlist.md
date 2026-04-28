# watchlist

Manage saved watchlist configurations and watchlist ticker summaries.

| Command | Use when | Required | Output |
|---|---|---|---|
| `watchlist configs` | List watchlists | none | JSON, CSV, TSV |
| `watchlist tickers` | Get watchlist tickers and nearest signals | none | JSON, CSV, TSV |
| `watchlist delete` | Delete a watchlist | `--key` | JSON |
| `watchlist add-ticker` | Add one ticker | `--watchlist-key`, `--ticker` | JSON |
| `watchlist create` | Create a watchlist | `--name` | JSON |
| `watchlist edit` | Replace/update a watchlist | `--key` | JSON |

```bash
volumeleaders-agent watchlist configs --format csv
volumeleaders-agent watchlist tickers --watchlist-key 12345 --format tsv
volumeleaders-agent watchlist add-ticker --watchlist-key 12345 --ticker AAPL
volumeleaders-agent watchlist create --name "Tech Sweeps" --tickers "AAPL,MSFT,GOOGL" --no-dark-pools --min-dollars 1000000
volumeleaders-agent watchlist edit --key 12345 --name "Updated Watchlist"
volumeleaders-agent watchlist delete --key 12345
```

`watchlist configs` fields include `SearchTemplateKey` for `--key` or `--watchlist-key`, `Name`, `Tickers`, volume/price/dollar ranges, RSI filters, trade type booleans, `SecurityTypeKey`, `MinVCD`, `SectorIndustry`.

`watchlist tickers` optional flag: `--watchlist-key`, default `-1` means all watchlists. Fields: `Ticker`, `Price`, `NearestTop10TradeDate`, `NearestTop10TradeClusterDate`, `NearestTop10TradeLevel`.

Create/edit option groups:

| Category | Flags | Defaults |
|---|---|---|
| Tickers | `--tickers` | none, comma-separated, max 500 |
| Volume | `--min-volume`, `--max-volume` | 0, 2000000000 |
| Dollars | `--min-dollars`, `--max-dollars` | 0, 30000000000 |
| Price | `--min-price`, `--max-price` | 0, 100000 |
| Quality | `--min-vcd`, `--min-relative-size`, `--max-trade-rank` | 0, 0, -1 |
| Security | `--security-type`, `--sector-industry` | -1, none |
| Trade types | `--normal-prints`, `--signature-prints`, `--late-prints`, `--timely-prints`, `--dark-pools`, `--lit-exchanges`, `--sweeps`, `--blocks` | true, disable with `--no-{name}` |
| Sessions | `--premarket-trades`, `--rth-trades`, `--ah-trades`, `--opening-trades`, `--closing-trades`, `--phantom-trades`, `--offsetting-trades` | true, disable with `--no-{name}` |
| RSI | `--rsi-overbought-daily`, `--rsi-overbought-hourly`, `--rsi-oversold-daily`, `--rsi-oversold-hourly` | -1, where 1=yes, 0=no, -1=ignore |

Edit gotcha: unspecified edit flags reset to defaults, so include every value that must remain set.
