package models

// PresetInfo describes a trade filter preset for JSON output.
type PresetInfo struct {
	Name    string            `json:"Name"`
	Group   string            `json:"Group"`
	Filters map[string]string `json:"Filters"`
}
