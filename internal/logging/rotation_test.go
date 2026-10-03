package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

func TestNewWriterUsesOverrideAndRotationDefaults(t *testing.T) {
	override := filepath.Join(t.TempDir(), "custom.log")
	t.Setenv("NALA_LOG_FILE", override)

	writer := NewWriter(filepath.Join(t.TempDir(), "default.log"))
	if writer.Filename != override {
		t.Fatalf("Filename = %q, want %q", writer.Filename, override)
	}
	if writer.MaxSize != 10 || writer.MaxBackups != 5 || writer.MaxAge != 30 || !writer.Compress {
		t.Fatalf("unexpected rotation settings: %+v", writer)
	}
}

func TestNewWriterUsesDefaultPath(t *testing.T) {
	t.Setenv("NALA_LOG_FILE", " ")
	path := filepath.Join(t.TempDir(), "service.log")
	writer := NewWriter(path)
	if writer.Filename != path {
		t.Fatalf("Filename = %q, want %q", writer.Filename, path)
	}
}

func TestRotatingWriterCreatesBackupOnDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.log")
	writer := &lumberjack.Logger{Filename: path, MaxSize: 1}
	first := bytes.Repeat([]byte("x"), 1<<20)
	if n, err := writer.Write(first); err != nil || n != len(first) {
		t.Fatalf("write initial log: n=%d err=%v", n, err)
	}
	if _, err := writer.Write([]byte("current log\n")); err != nil {
		t.Fatalf("write after rotation threshold: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close rotating writer: %v", err)
	}

	backups, err := filepath.Glob(filepath.Join(filepath.Dir(path), "service-*.log"))
	if err != nil {
		t.Fatalf("find rotated log: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("rotated logs = %v, want one timestamped backup", backups)
	}
	rotated, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read rotated log: %v", err)
	}
	if !bytes.Equal(rotated, first) {
		t.Fatalf("rotated log length = %d, want %d", len(rotated), len(first))
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if string(current) != "current log\n" {
		t.Fatalf("current log = %q, want %q", current, "current log\n")
	}
}
