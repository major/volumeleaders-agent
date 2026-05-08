package common

import (
	vlgo "github.com/major/volumeleaders-go/volumeleaders"

	"github.com/major/volumeleaders-agent/internal/models"
)

// stringToPtr converts a string to a *string, returning nil for empty strings.
// This preserves null JSON output for fields that VolumeLeaders returns as null
// but volumeleaders-go deserializes into zero-value strings.
//
// Workaround for https://github.com/major/volumeleaders-go/issues/7
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ptrToString dereferences a *string, returning "" for nil pointers.
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// convertDate converts a vlgo AspNetDate to an internal models AspNetDate.
func convertDate(d vlgo.AspNetDate) models.AspNetDate {
	return models.AspNetDate{Time: d.Time, Valid: d.Valid}
}

// MapVLTrade converts a vlgo Trade to an internal models Trade.
//
// Handles TradeID: int64 -> int.
func MapVLTrade(src *vlgo.Trade) models.Trade {
	return models.Trade{
		Date:                          convertDate(src.Date),
		StartDate:                     convertDate(src.StartDate),
		EndDate:                       convertDate(src.EndDate),
		TD30:                          convertDate(src.TD30),
		TD90:                          convertDate(src.TD90),
		TD1CY:                         convertDate(src.TD1CY),
		DateKey:                       src.DateKey,
		TimeKey:                       src.TimeKey,
		SecurityKey:                   src.SecurityKey,
		TradeID:                       int(src.TradeID),
		SequenceNumber:                src.SequenceNumber,
		EOM:                           models.FlexBool(src.EOM),
		EOQ:                           models.FlexBool(src.EOQ),
		EOY:                           models.FlexBool(src.EOY),
		OPEX:                          models.FlexBool(src.OPEX),
		VOLEX:                         models.FlexBool(src.VOLEX),
		Ticker:                        src.Ticker,
		Sector:                        src.Sector,
		Industry:                      src.Industry,
		Name:                          src.Name,
		FullDateTime:                  src.FullDateTime,
		FullTimeString24:              src.FullTimeString24,
		Price:                         src.Price,
		Bid:                           src.Bid,
		Ask:                           src.Ask,
		Dollars:                       src.Dollars,
		AverageBlockSizeDollars:       src.AverageBlockSizeDollars,
		AverageBlockSizeShares:        src.AverageBlockSizeShares,
		DollarsMultiplier:             src.DollarsMultiplier,
		Volume:                        src.Volume,
		AverageDailyVolume:            src.AverageDailyVolume,
		PercentDailyVolume:            src.PercentDailyVolume,
		RelativeSize:                  src.RelativeSize,
		LastComparibleTradeDate:       convertDate(src.LastComparibleTradeDate),
		IPODate:                       convertDate(src.IPODate),
		OffsettingTradeDate:           convertDate(src.OffsettingTradeDate),
		PhantomPrintFulfillmentDate:   convertDate(src.PhantomPrintFulfillmentDate),
		PhantomPrintFulfillmentDays:   src.PhantomPrintFulfillmentDays,
		TradeCount:                    src.TradeCount,
		CumulativeDistribution:        src.CumulativeDistribution,
		TradeRank:                     src.TradeRank,
		TradeRankSnapshot:             src.TradeRankSnapshot,
		LatePrint:                     models.FlexBool(src.LatePrint),
		Sweep:                         models.FlexBool(src.Sweep),
		DarkPool:                      models.FlexBool(src.DarkPool),
		OpeningTrade:                  models.FlexBool(src.OpeningTrade),
		ClosingTrade:                  models.FlexBool(src.ClosingTrade),
		PhantomPrint:                  models.FlexBool(src.PhantomPrint),
		InsideBar:                     models.FlexBool(src.InsideBar),
		DoubleInsideBar:               models.FlexBool(src.DoubleInsideBar),
		SignaturePrint:                models.FlexBool(src.SignaturePrint),
		NewPosition:                   models.FlexBool(src.NewPosition),
		AHInstitutionalDollars:        src.AHInstitutionalDollars,
		AHInstitutionalDollarsRank:    src.AHInstitutionalDollarsRank,
		AHInstitutionalVolume:         src.AHInstitutionalVolume,
		TotalInstitutionalDollars:     src.TotalInstitutionalDollars,
		TotalInstitutionalDollarsRank: src.TotalInstitutionalDollarsRank,
		TotalInstitutionalVolume:      src.TotalInstitutionalVolume,
		ClosingTradeDollars:           src.ClosingTradeDollars,
		ClosingTradeDollarsRank:       src.ClosingTradeDollarsRank,
		ClosingTradeVolume:            src.ClosingTradeVolume,
		TotalDollars:                  src.TotalDollars,
		TotalDollarsRank:              src.TotalDollarsRank,
		TotalVolume:                   src.TotalVolume,
		ClosePrice:                    src.ClosePrice,
		RSIHour:                       src.RSIHour,
		RSIDay:                        src.RSIDay,
		TotalRows:                     src.TotalRows,
		TradeConditions:               src.TradeConditions,
		FrequencyLast30TD:             src.FrequencyLast30TD,
		FrequencyLast90TD:             src.FrequencyLast90TD,
		FrequencyLast1CY:              src.FrequencyLast1CY,
		Cancelled:                     models.FlexBool(src.Cancelled), //nolint:misspell // matches VolumeLeaders API field name
		TotalTrades:                   src.TotalTrades,
		ExternalFeed:                  models.FlexBool(src.ExternalFeed),
	}
}

// MapVLTradeCluster converts a vlgo TradeCluster to an internal models
// TradeCluster. Also serves TradeClusterAlert since both are type aliases.
func MapVLTradeCluster(src *vlgo.TradeCluster) models.TradeCluster {
	return models.TradeCluster{
		Date:                           convertDate(src.Date),
		DateKey:                        src.DateKey,
		SecurityKey:                    src.SecurityKey,
		Ticker:                         src.Ticker,
		Sector:                         src.Sector,
		Industry:                       stringToPtr(src.Industry),
		Name:                           src.Name,
		MinFullDateTime:                src.MinFullDateTime,
		MaxFullDateTime:                src.MaxFullDateTime,
		MinFullTimeString24:            src.MinFullTimeString24,
		MaxFullTimeString24:            src.MaxFullTimeString24,
		ClosePrice:                     src.ClosePrice,
		Price:                          src.Price,
		Dollars:                        src.Dollars,
		AverageBlockSizeShares:         src.AverageBlockSizeShares,
		AverageBlockSizeDollars:        src.AverageBlockSizeDollars,
		Volume:                         src.Volume,
		TradeCount:                     src.TradeCount,
		IPODate:                        convertDate(src.IPODate),
		DollarsMultiplier:              src.DollarsMultiplier,
		CumulativeDistribution:         src.CumulativeDistribution,
		AverageDailyVolume:             src.AverageDailyVolume,
		EOM:                            models.FlexBool(src.EOM),
		EOQ:                            models.FlexBool(src.EOQ),
		EOY:                            models.FlexBool(src.EOY),
		OPEX:                           models.FlexBool(src.OPEX),
		VOLEX:                          models.FlexBool(src.VOLEX),
		InsideBar:                      models.FlexBool(src.InsideBar),
		DoubleInsideBar:                models.FlexBool(src.DoubleInsideBar),
		LastComparibleTradeClusterDate: convertDate(src.LastComparibleTradeClusterDate),
		TradeClusterRank:               src.TradeClusterRank,
		TotalRows:                      src.TotalRows,
		ExternalFeed:                   models.FlexBool(src.ExternalFeed),
	}
}

// MapVLTradeClusterBomb converts a vlgo TradeClusterBomb to an internal models
// TradeClusterBomb.
func MapVLTradeClusterBomb(src *vlgo.TradeClusterBomb) models.TradeClusterBomb {
	return models.TradeClusterBomb{
		Date:                               convertDate(src.Date),
		DateKey:                            src.DateKey,
		SecurityKey:                        src.SecurityKey,
		Ticker:                             src.Ticker,
		Sector:                             src.Sector,
		Industry:                           stringToPtr(src.Industry),
		Name:                               src.Name,
		MinFullDateTime:                    src.MinFullDateTime,
		MaxFullDateTime:                    src.MaxFullDateTime,
		MinFullTimeString24:                src.MinFullTimeString24,
		MaxFullTimeString24:                src.MaxFullTimeString24,
		ClosePrice:                         src.ClosePrice,
		Dollars:                            src.Dollars,
		AverageBlockSizeShares:             src.AverageBlockSizeShares,
		AverageBlockSizeDollars:            src.AverageBlockSizeDollars,
		Volume:                             src.Volume,
		TradeCount:                         src.TradeCount,
		IPODate:                            convertDate(src.IPODate),
		DollarsMultiplier:                  src.DollarsMultiplier,
		CumulativeDistribution:             src.CumulativeDistribution,
		AverageDailyVolume:                 src.AverageDailyVolume,
		EOM:                                models.FlexBool(src.EOM),
		EOQ:                                models.FlexBool(src.EOQ),
		EOY:                                models.FlexBool(src.EOY),
		OPEX:                               models.FlexBool(src.OPEX),
		VOLEX:                              models.FlexBool(src.VOLEX),
		InsideBar:                          models.FlexBool(src.InsideBar),
		DoubleInsideBar:                    models.FlexBool(src.DoubleInsideBar),
		LastComparableTradeClusterBombDate: convertDate(src.LastComparableTradeClusterBombDate),
		TradeClusterBombRank:               src.TradeClusterBombRank,
		TotalRows:                          src.TotalRows,
		ExternalFeed:                       models.FlexBool(src.ExternalFeed),
	}
}

// MapVLTradeLevel converts a vlgo TradeLevel to an internal models TradeLevel.
// The internal model has fewer fields; extra vlgo fields are dropped.
func MapVLTradeLevel(src *vlgo.TradeLevel) models.TradeLevel {
	return models.TradeLevel{
		Ticker:                 &src.Ticker,
		Name:                   src.Name,
		Price:                  src.Price,
		Dollars:                src.Dollars,
		Volume:                 src.Volume,
		Trades:                 src.Trades,
		RelativeSize:           src.RelativeSize,
		CumulativeDistribution: src.CumulativeDistribution,
		TradeLevelRank:         src.TradeLevelRank,
		MinDate:                convertDate(src.MinDate),
		MaxDate:                convertDate(src.MaxDate),
		Dates:                  src.Dates,
	}
}

// MapVLTradeLevelTouch converts a vlgo TradeLevel to an internal models
// TradeLevelTouch. Both endpoints return the same vlgo type but the CLI maps
// them to different internal models.
func MapVLTradeLevelTouch(src *vlgo.TradeLevel) models.TradeLevelTouch {
	return models.TradeLevelTouch{
		Ticker:                 src.Ticker,
		Sector:                 src.Sector,
		Industry:               src.Industry,
		Name:                   ptrToString(src.Name),
		Date:                   convertDate(src.Date),
		MinDate:                convertDate(src.MinDate),
		MaxDate:                convertDate(src.MaxDate),
		FullDateTime:           ptrToString(src.FullDateTime),
		FullTimeString24:       src.FullTimeString24,
		Dates:                  src.Dates,
		Price:                  src.Price,
		Dollars:                src.Dollars,
		Volume:                 src.Volume,
		Trades:                 src.Trades,
		CumulativeDistribution: src.CumulativeDistribution,
		TradeLevelRank:         src.TradeLevelRank,
		TotalRows:              src.TotalRows,
		TradeLevelTouches:      src.TradeLevelTouches,
		RelativeSize:           src.RelativeSize,
	}
}

// MapVLAlertConfig converts a vlgo AlertConfig to an internal models
// AlertConfig. Both types are structurally identical.
func MapVLAlertConfig(src *vlgo.AlertConfig) models.AlertConfig {
	return models.AlertConfig{
		AlertConfigKey:         src.AlertConfigKey,
		UserKey:                src.UserKey,
		Name:                   src.Name,
		Tickers:                src.Tickers,
		TradeRankLTE:           src.TradeRankLTE,
		TradeVCDGTE:            src.TradeVCDGTE,
		TradeMultGTE:           src.TradeMultGTE,
		TradeVolumeGTE:         src.TradeVolumeGTE,
		TradeDollarsGTE:        src.TradeDollarsGTE,
		TradeConditions:        src.TradeConditions,
		TradeClusterRankLTE:    src.TradeClusterRankLTE,
		TradeClusterVCDGTE:     src.TradeClusterVCDGTE,
		TradeClusterMultGTE:    src.TradeClusterMultGTE,
		TradeClusterVolumeGTE:  src.TradeClusterVolumeGTE,
		TradeClusterDollarsGTE: src.TradeClusterDollarsGTE,
		TotalRankLTE:           src.TotalRankLTE,
		TotalVolumeGTE:         src.TotalVolumeGTE,
		TotalDollarsGTE:        src.TotalDollarsGTE,
		AHRankLTE:              src.AHRankLTE,
		AHVolumeGTE:            src.AHVolumeGTE,
		AHDollarsGTE:           src.AHDollarsGTE,
		ClosingTradeRankLTE:    src.ClosingTradeRankLTE,
		ClosingTradeVCDGTE:     src.ClosingTradeVCDGTE,
		ClosingTradeMultGTE:    src.ClosingTradeMultGTE,
		ClosingTradeVolumeGTE:  src.ClosingTradeVolumeGTE,
		ClosingTradeDollarsGTE: src.ClosingTradeDollarsGTE,
		ClosingTradeConditions: src.ClosingTradeConditions,
		OffsettingPrint:        src.OffsettingPrint,
		PhantomPrint:           src.PhantomPrint,
		Sweep:                  src.Sweep,
		DarkPool:               src.DarkPool,
	}
}

// MapVLTradeAlert converts a vlgo TradeAlert to an internal models TradeAlert.
//
// Handles TradeID: int64 -> int.
func MapVLTradeAlert(src *vlgo.TradeAlert) models.TradeAlert {
	return models.TradeAlert{
		Date:                         convertDate(src.Date),
		StartDate:                    convertDate(src.StartDate),
		EndDate:                      convertDate(src.EndDate),
		FullTimeString24:             src.FullTimeString24,
		DateKey:                      src.DateKey,
		SecurityKey:                  src.SecurityKey,
		TimeKey:                      src.TimeKey,
		TradeID:                      int(src.TradeID),
		SequenceNumber:               src.SequenceNumber,
		UserKey:                      src.UserKey,
		UserKeys:                     src.UserKeys,
		Sent:                         src.Sent,
		Email:                        src.Email,
		Emails:                       src.Emails,
		Ticker:                       src.Ticker,
		Sector:                       src.Sector,
		Industry:                     src.Industry,
		Name:                         src.Name,
		AlertType:                    src.AlertType,
		Price:                        src.Price,
		TradeRank:                    src.TradeRank,
		VolumeCumulativeDistribution: src.VolumeCumulativeDistribution,
		DollarsMultiplier:            src.DollarsMultiplier,
		Volume:                       src.Volume,
		Dollars:                      src.Dollars,
		LastComparibleTradeDateKey:   src.LastComparibleTradeDateKey,
		LastComparibleTradeDate:      convertDate(src.LastComparibleTradeDate),
		OffsettingTradeDate:          convertDate(src.OffsettingTradeDate),
		PhantomPrintFulfillmentDate:  convertDate(src.PhantomPrintFulfillmentDate),
		FullDateTime:                 src.FullDateTime,
		IPODate:                      convertDate(src.IPODate),
		RSIHour:                      src.RSIHour,
		RSIDay:                       src.RSIDay,
		InProcess:                    src.InProcess,
		Complete:                     src.Complete,
		Sweep:                        models.FlexBool(src.Sweep),
		DarkPool:                     models.FlexBool(src.DarkPool),
		LatePrint:                    models.FlexBool(src.LatePrint),
		ClosingTrade:                 models.FlexBool(src.ClosingTrade),
		SignaturePrint:               models.FlexBool(src.SignaturePrint),
		PhantomPrint:                 models.FlexBool(src.PhantomPrint),
	}
}

// MapVLEarning converts a vlgo Earning to an internal models Earnings.
// Extra vlgo fields (Date, Current, TotalRows) are dropped.
func MapVLEarning(src *vlgo.Earning) models.Earnings {
	return models.Earnings{
		Ticker:                src.Ticker,
		Name:                  src.Name,
		Sector:                src.Sector,
		Industry:              src.Industry,
		EarningsDate:          convertDate(src.EarningsDate),
		AfterMarketClose:      src.AfterMarketClose,
		TradeCount:            src.TradeCount,
		TradeClusterCount:     src.TradeClusterCount,
		TradeClusterBombCount: src.TradeClusterBombCount,
	}
}

// MapVLExhaustionScores converts a vlgo ExhaustionScores to an internal models
// ExhaustionScore.
func MapVLExhaustionScores(src vlgo.ExhaustionScores) models.ExhaustionScore {
	return models.ExhaustionScore{
		DateKey:                   src.DateKey,
		ExhaustionScoreRank:       src.ExhaustionScoreRank,
		ExhaustionScoreRank30Day:  src.ExhaustionScoreRank30Day,
		ExhaustionScoreRank90Day:  src.ExhaustionScoreRank90Day,
		ExhaustionScoreRank365Day: src.ExhaustionScoreRank365Day,
	}
}
