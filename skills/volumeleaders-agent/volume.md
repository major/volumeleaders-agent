# volume

Volume leaderboards ranking stocks by trading activity. All commands require `--date`, support `--tickers` with aliases (`--ticker`, `--symbol`, `--symbols`) or positional tickers, `--format json|csv|tsv`, and pagination. Default sort is `--order-dir asc`, unlike trade commands.

| Command | Use when |
|---|---|
| `volume institutional` | Rank stocks by institutional dollar volume |
| `volume ah-institutional` | Rank after-hours institutional activity |
| `volume total` | Rank total market volume across trade types |

```bash
volumeleaders-agent volume institutional --date 2026-04-28
volumeleaders-agent volume institutional XLE XLK --date 2026-04-28
volumeleaders-agent volume institutional --date 2026-04-28 --format csv
volumeleaders-agent volume ah-institutional --date 2026-04-28
volumeleaders-agent volume total --date 2026-04-28
```

Output model: same Trade model as `trade list`. Key volume fields: `Ticker`, `Name`, `Sector`, `Industry`, `TotalInstitutionalDollars`, `TotalInstitutionalDollarsRank`, `TotalInstitutionalVolume`, `AHInstitutionalDollars`, `AHInstitutionalDollarsRank`, `AHInstitutionalVolume`, `ClosingTradeDollars`, `ClosingTradeDollarsRank`, `ClosingTradeVolume`, `TotalDollars`, `TotalDollarsRank`, `TotalVolume`, `AverageDailyVolume`, `PercentDailyVolume`, `ClosePrice`, `RSIHour`, `RSIDay`.
