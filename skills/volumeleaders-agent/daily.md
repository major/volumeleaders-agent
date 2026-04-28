# daily

Compact cross-endpoint daily summaries for institutional activity. 1 subcommand. Use this first when you need a fast market-wide read before drilling into individual trade, volume, cluster, or level-touch commands.

## daily summary

Summarize one trading day across institutional volume, trade clusters, cluster bombs, level touches, leveraged ETF sentiment, sector totals, and market exhaustion. Output is compact JSON only because the response is a nested summary rather than a list-style table.

Required: `--date`
Optional: `--limit` (default `10`, maximum rows per leaderboard section)

```bash
volumeleaders-agent daily summary --date 2026-04-28
volumeleaders-agent --pretty daily summary --date 2026-04-28 --limit 5
```

Output sections:

- `date`: requested trading date.
- `top_institutional_volume_tickers`: largest total institutional dollar volume tickers from `volume institutional` data.
- `top_clusters_by_dollars`: largest trade clusters by dollar value.
- `top_clusters_by_multiplier`: most unusual trade clusters by dollar multiplier.
- `repeated_cluster_tickers`: tickers with two or more clusters on the requested day.
- `sector_totals`: sector-level totals merged across institutional volume, clusters, cluster bombs, and level touches.
- `cluster_bombs`: largest sudden aggressive cluster bursts.
- `level_touches`: notable level-touch events split into `by_relative_size` and `by_dollars` leaderboards.
- `leveraged_etf_sentiment`: leveraged ETF bull/bear flow using the same classifier and defaults as `trade sentiment`.
- `market_exhaustion`: market exhaustion ranks from `market exhaustion` for the requested date.

Data sources queried by this command:

- `/InstitutionalVolume/GetInstitutionalVolume`
- `/TradeClusters/GetTradeClusters`
- `/TradeClusterBombs/GetTradeClusterBombs`
- `/TradeLevelTouches/GetTradeLevelTouches`
- `/Trades/GetTrades` with the leveraged ETF combined sector filter
- `/ExecutiveSummary/GetExhaustionScores`

Use follow-up commands for full detail:

```bash
# Drill into a ticker from the daily summary
volumeleaders-agent trade list --tickers NVDA --start-date 2026-04-28 --end-date 2026-04-28

# Inspect the full cluster table for the same day
volumeleaders-agent trade clusters --start-date 2026-04-28 --end-date 2026-04-28 --length -1

# Re-run leveraged ETF sentiment over a wider window
volumeleaders-agent trade sentiment --start-date 2026-04-22 --end-date 2026-04-28
```
