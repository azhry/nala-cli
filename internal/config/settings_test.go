package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSettingsStoreSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	store := &SettingsStore{Path: path}
	want := Settings{APIBaseURL: "http://127.0.0.1:8082", SVCBaseURL: "http://127.0.0.1:8083"}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat config file: %v", err)
		}
		if permission := info.Mode().Perm(); permission != 0o600 {
			t.Fatalf("config file permissions = %o, want 600", permission)
		}
	}
}

func TestSettingsStoreMissingFileIsEmpty(t *testing.T) {
	store := &SettingsStore{Path: filepath.Join(t.TempDir(), "missing.json")}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != (Settings{}) {
		t.Fatalf("Load() = %+v, want empty settings", got)
	}
}

func TestSettingsStoreRejectsInvalidBaseURL(t *testing.T) {
	store := &SettingsStore{Path: filepath.Join(t.TempDir(), "config.json")}
	if err := store.Save(Settings{APIBaseURL: "not-a-url"}); err == nil {
		t.Fatal("Save() accepted an invalid API URL")
	}
}
