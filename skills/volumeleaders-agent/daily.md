# daily

Use `daily summary` first when the user wants a fast market-wide institutional activity read before drilling into trades, volume, clusters, or levels.

## daily summary

Compact nested JSON-only summary for one trading day. The schema is optimized for LLM stock-analysis workflows by removing duplicate leaderboards and retaining why each signal was included.

Required: `--date`. Optional: `--limit` rows considered per ranking, default `10`. Merged sections such as `clusters.top` and `level_touches` can return more than `--limit` rows when different rankings select different signals.

```bash
volumeleaders-agent daily summary --date 2026-04-28
volumeleaders-agent --pretty daily summary --date 2026-04-28 --limit 5
```

Output sections: `date`, `institutional_volume`, `clusters`, `cluster_bombs`, `level_touches`, `leveraged_etf_sentiment`, `market_exhaustion`.

- `institutional_volume` lists top tickers by institutional dollars with `ticker`, optional `sector`, `price`, `institutional_dollars`, and `rank`.
- `clusters.top` is a deduped union of top clusters by dollars and multiplier. Check `top_by` for why a row was selected, usually `dollars`, `multiplier`, or both.
- `clusters.repeated_tickers` groups tickers with multiple clusters and includes total cluster dollars, trade count, max multiplier, and best rank.
- `cluster_bombs` keeps sudden burst signals with dollars, multiplier, trade count, rank, and cumulative distribution.
- `level_touches` is a deduped union of top touches by relative size and dollars. Check `top_by` for the ranking reason. Rows include price, dollars, relative size, trades, rank, touches, and cumulative distribution.
- `leveraged_etf_sentiment` is flattened to signal, ratio, bull/bear dollars, bull/bear trades, and top bull/bear tickers. Treat it as leveraged ETF proxy flow, not signed buy/sell flow.
- `market_exhaustion` returns compact exhaustion ranks as `rank`, `rank_30d`, `rank_90d`, and `rank_365d`. Lower ranks indicate stronger exhaustion or reversal signal.

The summary intentionally omits mixed sector totals because combining dollars and volume across endpoint summaries can double count the same underlying activity.

Data sources: institutional volume, trade clusters, cluster bombs, level touches, leveraged ETF sentiment via one capped 50-row `Trades/GetTrades` request, and exhaustion scores.

Follow up with:

```bash
volumeleaders-agent trade list --tickers NVDA --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent trade clusters --start-date 2026-04-28 --end-date 2026-04-28 --length -1
volumeleaders-agent trade sentiment --start-date 2026-04-22 --end-date 2026-04-28
```
