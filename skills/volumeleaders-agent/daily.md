# daily

Use `daily summary` first when the user wants a fast market-wide institutional activity read before drilling into trades, volume, clusters, or levels.

## daily summary

Nested JSON-only summary for one trading day.

Required: `--date`. Optional: `--limit` leaderboard row limit, default `10`.

```bash
volumeleaders-agent daily summary --date 2026-04-28
volumeleaders-agent --pretty daily summary --date 2026-04-28 --limit 5
```

Output sections: `date`, `top_institutional_volume_tickers`, `top_clusters_by_dollars`, `top_clusters_by_multiplier`, `repeated_cluster_tickers`, `sector_totals`, `cluster_bombs`, `level_touches` split by relative size and dollars, `leveraged_etf_sentiment`, `market_exhaustion`.

Data sources: institutional volume, trade clusters, cluster bombs, level touches, leveraged ETF sentiment via `Trades/GetTrades`, and exhaustion scores.

Follow up with:

```bash
volumeleaders-agent trade list --tickers NVDA --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent trade clusters --start-date 2026-04-28 --end-date 2026-04-28 --length -1
volumeleaders-agent trade sentiment --start-date 2026-04-22 --end-date 2026-04-28
```
