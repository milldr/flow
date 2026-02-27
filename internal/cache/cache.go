// Package cache provides simple file-backed caching for flow.
package cache

import (
	"encoding/json"
	"os"
	"time"
)

// StatusEntry holds a cached status for a single workspace.
type StatusEntry struct {
	Status     string    `json:"status"`
	ResolvedAt time.Time `json:"resolved_at"`
}

// StatusCache maps workspace IDs to their cached status.
type StatusCache map[string]StatusEntry

// LoadStatus reads the status cache from disk. Returns an empty cache if
// the file doesn't exist or can't be parsed.
func LoadStatus(path string) StatusCache {
	data, err := os.ReadFile(path)
	if err != nil {
		return make(StatusCache)
	}
	var c StatusCache
	if err := json.Unmarshal(data, &c); err != nil {
		return make(StatusCache)
	}
	return c
}

// SaveStatus writes the status cache to disk as JSON.
func SaveStatus(path string, c StatusCache) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
