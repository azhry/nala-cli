package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWithLoggingPreservesHelpOutput(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "nala-cli.log")
	t.Setenv("NALA_LOG_FILE", logPath)

	var stdout, stderr bytes.Buffer
	if err := runWithLogging([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("run help: %v", err)
	}
	if stdout.String() != "Usage:\n  nala login\n  nala user info\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read CLI log: %v", err)
	}
	if !bytes.Contains(log, []byte("command completed")) || !bytes.Contains(log, []byte("help")) {
		t.Fatalf("log does not describe help completion: %s", log)
	}
}

func TestRunWithLoggingDoesNotRecordArgumentsOrErrorText(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "nala-cli.log")
	t.Setenv("NALA_LOG_FILE", logPath)
	secretArgument := "user-supplied-private-value"

	err := runWithLogging([]string{secretArgument}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected unknown-command error")
	}
	log, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("read CLI log: %v", readErr)
	}
	if !bytes.Contains(log, []byte("command failed")) {
		t.Fatalf("log does not describe failure: %s", log)
	}
	if bytes.Contains(log, []byte(secretArgument)) || bytes.Contains(log, []byte(err.Error())) {
		t.Fatalf("log contains command arguments or error text: %s", log)
	}
	if !bytes.Contains(log, []byte(fmt.Sprintf("%T", err))) {
		t.Fatalf("log does not include the error type: %s", log)
	}
}

func TestCLILogPathUsesConfigDirectoryAndProcessID(t *testing.T) {
	root := t.TempDir()
	t.Setenv("NALA_CONFIG_DIR", root)

	path, err := cliLogPath()
	if err != nil {
		t.Fatalf("resolve CLI log path: %v", err)
	}
	want := filepath.Join(root, "logs", fmt.Sprintf("nala-cli-%d.log", os.Getpid()))
	if path != want {
		t.Fatalf("log path = %q, want %q", path, want)
	}
}
