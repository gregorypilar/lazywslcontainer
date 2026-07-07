package client

import "context"

type Stat struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
	MemPerc  string `json:"MemPerc"`
	NetIO    string `json:"NetIO"`
	BlockIO  string `json:"BlockIO"`
	PIDs     int    `json:"PIDs"`
}

func (w *WSLC) Stats(ctx context.Context) ([]Stat, error) {
	var stats []Stat
	err := w.runJSON(ctx, &stats, "stats", "--format", "json")
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (w *WSLC) StatsRaw(ctx context.Context) ([]byte, error) {
	return w.run(ctx, "stats", "--format", "json")
}
