package watchlist

import (
	"context"
	"fmt"
	"net/url"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/models"
)

type watchlistService interface {
	Configs(context.Context, common.DataTableOptions) ([]models.WatchListConfig, error)
	Tickers(context.Context, common.DataTableOptions) ([]models.WatchListTicker, error)
	Delete(ctx context.Context, key int) error
	AddTicker(ctx context.Context, watchListKey int, ticker string) (*vlgo.AddTickerToWatchListResponse, error)
	SaveConfig(ctx context.Context, fields map[string]string) error
}

type volumeLeadersWatchlistService struct {
	client *vlgo.Client
}

func newWatchlistService(ctx context.Context) (watchlistService, error) {
	commandClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	vlClient, err := commandClient.NewVolumeLeadersClient()
	if err != nil {
		return nil, err
	}
	return &volumeLeadersWatchlistService{client: vlClient}, nil
}

func (svc *volumeLeadersWatchlistService) Configs(ctx context.Context, opts common.DataTableOptions) ([]models.WatchListConfig, error) {
	if opts.Length < 0 {
		return svc.fetchAllConfigs(ctx, opts)
	}
	resp, err := svc.client.GetWatchLists(ctx, vlgo.WatchListsRequest{DataTables: newVolumeLeadersDataTablesRequest(opts)})
	if err != nil {
		return nil, fmt.Errorf("query watchlist configs: %w", err)
	}
	return mapWatchListConfigs(resp.Data), nil
}

func (svc *volumeLeadersWatchlistService) fetchAllConfigs(ctx context.Context, opts common.DataTableOptions) ([]models.WatchListConfig, error) {
	pageSize := common.PaginationPageSize
	initialStart := opts.Start
	all := make([]models.WatchListConfig, 0)
	for {
		opts.Length = pageSize
		resp, err := svc.client.GetWatchLists(ctx, vlgo.WatchListsRequest{DataTables: newVolumeLeadersDataTablesRequest(opts)})
		if err != nil {
			return nil, fmt.Errorf("query watchlist configs: %w", err)
		}
		page := mapWatchListConfigs(resp.Data)
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		if fetchedAllRecords(initialStart, len(all), resp.RecordsFiltered) || len(page) < pageSize {
			break
		}
		opts.Start += len(page)
	}
	return all, nil
}

func newVolumeLeadersDataTablesRequest(opts common.DataTableOptions) vlgo.DataTablesRequest {
	return vlgo.DataTablesRequest{
		Draw:   1,
		Start:  opts.Start,
		Length: opts.Length,
		Order: []vlgo.DataTablesOrder{{
			Column: opts.OrderCol,
			Dir:    string(opts.OrderDir),
		}},
	}
}

func fetchedAllRecords(initialStart, fetched, recordsFiltered int) bool {
	return recordsFiltered > 0 && initialStart+fetched >= recordsFiltered
}

func mapWatchListConfigs(configs []vlgo.WatchListConfig) []models.WatchListConfig {
	mapped := make([]models.WatchListConfig, 0, len(configs))
	for i := range configs {
		mapped = append(mapped, mapWatchListConfig(&configs[i]))
	}
	return mapped
}

func mapWatchListConfig(config *vlgo.WatchListConfig) models.WatchListConfig {
	return models.WatchListConfig{
		SearchTemplateKey:           config.SearchTemplateKey,
		UserKey:                     config.UserKey,
		SearchTemplateTypeKey:       config.SearchTemplateTypeKey,
		Name:                        config.Name,
		Tickers:                     config.Tickers,
		SortOrder:                   config.SortOrder,
		MinVolume:                   config.MinVolume,
		MaxVolume:                   config.MaxVolume,
		MinDollars:                  config.MinDollars,
		MaxDollars:                  config.MaxDollars,
		MinPrice:                    config.MinPrice,
		MaxPrice:                    config.MaxPrice,
		RSIOverboughtHourly:         config.RSIOverboughtHourly,
		RSIOverboughtDaily:          config.RSIOverboughtDaily,
		RSIOversoldHourly:           config.RSIOversoldHourly,
		RSIOversoldDaily:            config.RSIOversoldDaily,
		Conditions:                  config.Conditions,
		RSIOverboughtHourlySelected: config.RSIOverboughtHourlySelected,
		RSIOverboughtDailySelected:  config.RSIOverboughtDailySelected,
		RSIOversoldHourlySelected:   config.RSIOversoldHourlySelected,
		RSIOversoldDailySelected:    config.RSIOversoldDailySelected,
		MinRelativeSize:             config.MinRelativeSize,
		MinRelativeSizeSelected:     config.MinRelativeSizeSelected,
		MaxTradeRank:                config.MaxTradeRank,
		SecurityTypeKey:             config.SecurityTypeKey,
		SecurityType:                config.SecurityType,
		MaxTradeRankSelected:        config.MaxTradeRankSelected,
		MinVCD:                      config.MinVCD,
		NormalPrints:                config.NormalPrints,
		SignaturePrints:             config.SignaturePrints,
		LatePrints:                  config.LatePrints,
		TimelyPrints:                config.TimelyPrints,
		DarkPools:                   config.DarkPools,
		LitExchanges:                config.LitExchanges,
		Sweeps:                      config.Sweeps,
		Blocks:                      config.Blocks,
		PremarketTrades:             config.PremarketTrades,
		RTHTrades:                   config.RTHTrades,
		AHTrades:                    config.AHTrades,
		OpeningTrades:               config.OpeningTrades,
		ClosingTrades:               config.ClosingTrades,
		PhantomTrades:               config.PhantomTrades,
		OffsettingTrades:            config.OffsettingTrades,
		NormalPrintsSelected:        config.NormalPrintsSelected,
		SignaturePrintsSelected:     config.SignaturePrintsSelected,
		LatePrintsSelected:          config.LatePrintsSelected,
		TimelyPrintsSelected:        config.TimelyPrintsSelected,
		DarkPoolsSelected:           config.DarkPoolsSelected,
		LitExchangesSelected:        config.LitExchangesSelected,
		SweepsSelected:              config.SweepsSelected,
		BlocksSelected:              config.BlocksSelected,
		PremarketTradesSelected:     config.PremarketTradesSelected,
		RTHTradesSelected:           config.RTHTradesSelected,
		AHTradesSelected:            config.AHTradesSelected,
		OpeningTradesSelected:       config.OpeningTradesSelected,
		ClosingTradesSelected:       config.ClosingTradesSelected,
		PhantomTradesSelected:       config.PhantomTradesSelected,
		OffsettingTradesSelected:    config.OffsettingTradesSelected,
		SectorIndustry:              config.SectorIndustry,
		// APIKey is intentionally omitted to avoid leaking credentials to stdout.
	}
}

// Delete removes a watchlist by key.
func (svc *volumeLeadersWatchlistService) Delete(ctx context.Context, key int) error {
	if err := svc.client.DeleteWatchList(ctx, vlgo.DeleteWatchListRequest{WatchListKey: key}); err != nil {
		return fmt.Errorf("delete watchlist: %w", err)
	}
	return nil
}

// AddTicker adds a ticker symbol to an existing watchlist.
func (svc *volumeLeadersWatchlistService) AddTicker(ctx context.Context, watchListKey int, ticker string) (*vlgo.AddTickerToWatchListResponse, error) {
	resp, err := svc.client.AddTickerToWatchList(ctx, vlgo.AddTickerToWatchListRequest{
		WatchListKey: watchListKey,
		Ticker:       ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("add ticker to watchlist: %w", err)
	}
	return resp, nil
}

// Tickers returns tickers for a watchlist, paginating when opts.Length < 0.
func (svc *volumeLeadersWatchlistService) Tickers(ctx context.Context, opts common.DataTableOptions) ([]models.WatchListTicker, error) {
	if opts.Length < 0 {
		return svc.fetchAllTickers(ctx, opts)
	}
	resp, err := svc.client.GetWatchListTickers(ctx, vlgo.WatchListTickersRequest{
		DataTables: newVolumeLeadersDataTablesRequest(opts),
		Filters:    filtersToValues(opts.Filters),
	})
	if err != nil {
		return nil, fmt.Errorf("query watchlist tickers: %w", err)
	}
	return mapWatchListTickers(resp.Data), nil
}

func (svc *volumeLeadersWatchlistService) fetchAllTickers(ctx context.Context, opts common.DataTableOptions) ([]models.WatchListTicker, error) {
	pageSize := common.PaginationPageSize
	initialStart := opts.Start
	all := make([]models.WatchListTicker, 0)
	for {
		opts.Length = pageSize
		resp, err := svc.client.GetWatchListTickers(ctx, vlgo.WatchListTickersRequest{
			DataTables: newVolumeLeadersDataTablesRequest(opts),
			Filters:    filtersToValues(opts.Filters),
		})
		if err != nil {
			return nil, fmt.Errorf("query watchlist tickers: %w", err)
		}
		page := mapWatchListTickers(resp.Data)
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		if fetchedAllRecords(initialStart, len(all), resp.RecordsFiltered) || len(page) < pageSize {
			break
		}
		opts.Start += len(page)
	}
	return all, nil
}

// SaveConfig creates or edits a watchlist configuration.
func (svc *volumeLeadersWatchlistService) SaveConfig(ctx context.Context, fields map[string]string) error {
	values := make(url.Values, len(fields))
	for k, v := range fields {
		values.Set(k, v)
	}
	if err := svc.client.SaveWatchListConfig(ctx, vlgo.SaveWatchListConfigRequest{Fields: values}); err != nil {
		return fmt.Errorf("save watchlist config: %w", err)
	}
	return nil
}

// filtersToValues converts a string map to url.Values for vlgo request filters.
func filtersToValues(filters map[string]string) url.Values {
	if len(filters) == 0 {
		return nil
	}
	values := make(url.Values, len(filters))
	for k, v := range filters {
		values.Set(k, v)
	}
	return values
}

func mapWatchListTickers(tickers []vlgo.WatchListTicker) []models.WatchListTicker {
	mapped := make([]models.WatchListTicker, 0, len(tickers))
	for i := range tickers {
		mapped = append(mapped, mapWatchListTicker(&tickers[i]))
	}
	return mapped
}

func mapWatchListTicker(t *vlgo.WatchListTicker) models.WatchListTicker {
	var price float64
	if t.Price != nil {
		price = *t.Price
	}
	// vlgo exposes the nearest trade level as NearestTop10TradeLevelPrice;
	// map it to the models field NearestTop10TradeLevel, using nil when the
	// associated date is absent (no trade level data for this ticker).
	var nearestLevel *float64
	if t.NearestTop10TradeLevelDate.Valid {
		nearestLevel = &t.NearestTop10TradeLevelPrice
	}
	return models.WatchListTicker{
		Ticker:                       t.Ticker,
		Price:                        price,
		NearestTop10TradeDate:        models.AspNetDate{Time: t.NearestTop10TradeDate.Time, Valid: t.NearestTop10TradeDate.Valid},
		NearestTop10TradeClusterDate: models.AspNetDate{Time: t.NearestTop10TradeClusterDate.Time, Valid: t.NearestTop10TradeClusterDate.Valid},
		NearestTop10TradeLevel:       nearestLevel,
	}
}
