package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadStatusMissingFile(t *testing.T) {
	c := LoadStatus("/nonexistent/path/status.json")
	if len(c) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(c))
	}
}

func TestLoadStatusMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "status.json")
	if err := writeFile(path, "not json"); err != nil {
		t.Fatal(err)
	}
	c := LoadStatus(path)
	if len(c) != 0 {
		t.Errorf("expected empty cache for malformed JSON, got %d entries", len(c))
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "status.json")
	now := time.Now().Truncate(time.Second)

	original := StatusCache{
		"ws-1": {Status: "closed", ResolvedAt: now},
		"ws-2": {Status: "open", ResolvedAt: now},
	}

	if err := SaveStatus(path, original); err != nil {
		t.Fatalf("SaveStatus: %v", err)
	}

	loaded := LoadStatus(path)
	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded["ws-1"].Status != "closed" {
		t.Errorf("ws-1 status = %q, want \"closed\"", loaded["ws-1"].Status)
	}
	if loaded["ws-2"].Status != "open" {
		t.Errorf("ws-2 status = %q, want \"open\"", loaded["ws-2"].Status)
	}
	if !loaded["ws-1"].ResolvedAt.Equal(now) {
		t.Errorf("ws-1 resolved_at = %v, want %v", loaded["ws-1"].ResolvedAt, now)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "status.json")

	first := StatusCache{"ws-1": {Status: "open"}}
	if err := SaveStatus(path, first); err != nil {
		t.Fatal(err)
	}

	second := StatusCache{"ws-1": {Status: "closed"}}
	if err := SaveStatus(path, second); err != nil {
		t.Fatal(err)
	}

	loaded := LoadStatus(path)
	if loaded["ws-1"].Status != "closed" {
		t.Errorf("expected overwritten status \"closed\", got %q", loaded["ws-1"].Status)
	}
}

func TestSaveBadPath(t *testing.T) {
	err := SaveStatus("/nonexistent/dir/status.json", StatusCache{})
	if err == nil {
		t.Error("expected error writing to nonexistent directory")
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
