package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSessionStoreSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "session.json")
	store := &SessionStore{Path: path}
	if err := store.Save("session-token-for-test"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	token, err := store.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if token != "session-token-for-test" {
		t.Fatalf("Token() = %q", token)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat session file: %v", err)
		}
		if permission := info.Mode().Perm(); permission != 0o600 {
			t.Fatalf("session file permissions = %o, want 600", permission)
		}
	}
}

func TestSessionStoreRejectsEmptyToken(t *testing.T) {
	store := &SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save(" "); err == nil {
		t.Fatal("Save() accepted an empty token")
	}
}
