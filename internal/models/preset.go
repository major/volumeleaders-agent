package models

// PresetInfo describes a trade filter preset for JSON output.
type PresetInfo struct {
	Name    string            `json:"Name"`
	Group   string            `json:"Group"`
	Filters map[string]string `json:"Filters"`
}

// PresetTickersInfo describes the ticker symbols associated with a preset.
// Presets that use explicit ticker lists have Type "tickers" with a populated
// Tickers slice. Presets that filter by sector have Type "sector-filter" with
// a SectorIndustry value. Presets with neither have Type "unfiltered".
type PresetTickersInfo struct {
	Preset         string   `json:"Preset"`
	Group          string   `json:"Group"`
	Type           string   `json:"Type"`
	Tickers        []string `json:"Tickers,omitempty"`
	SectorIndustry string   `json:"SectorIndustry,omitempty"`
}
