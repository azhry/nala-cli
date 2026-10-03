package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/azhry/nala-cli/internal/auth"
	"github.com/azhry/nala-cli/internal/logging"
)

func main() {
	if err := runWithLogging(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "nala: %s\n", err)
		os.Exit(1)
	}
}

func runWithLogging(args []string, stdout, stderr io.Writer) error {
	path, err := cliLogPath()
	if err != nil {
		path = strings.TrimSpace(os.Getenv("NALA_LOG_FILE"))
		if path == "" {
			return run(args, stdout, stderr)
		}
	}

	rotatingWriter := logging.NewWriter(path)
	defer rotatingWriter.Close()
	logger := slog.New(slog.NewJSONHandler(rotatingWriter, nil))

	err = run(args, stdout, stderr)
	if err != nil {
		logger.Error("command failed", "command", safeCommandName(args), "error_type", fmt.Sprintf("%T", err))
		return err
	}
	logger.Info("command completed", "command", safeCommandName(args))
	return nil
}

func cliLogPath() (string, error) {
	root := strings.TrimSpace(os.Getenv("NALA_CONFIG_DIR"))
	if root == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(userConfigDir, "nala")
	}
	return filepath.Join(root, "logs", fmt.Sprintf("nala-cli-%d.log", os.Getpid())), nil
}

func safeCommandName(args []string) string {
	switch {
	case len(args) == 1 && args[0] == "login":
		return "login"
	case len(args) == 2 && args[0] == "user" && args[1] == "info":
		return "user info"
	case len(args) == 0 || len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		return "help"
	default:
		return "unknown"
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--help" || args[0] == "-h")) {
		writeUsage(stdout)
		return nil
	}
	switch {
	case len(args) == 1 && args[0] == "login":
		client, err := auth.NewClientFromEnvironment()
		if err != nil {
			return err
		}
		user, err := client.Login(context.Background())
		if err != nil {
			return err
		}
		name := user.Name
		if name == "" {
			name = user.ID
		}
		fmt.Fprintf(stdout, "Logged in as %s\n", name)
		return nil
	case len(args) == 2 && args[0] == "user" && args[1] == "info":
		client, err := auth.NewClientFromEnvironment()
		if err != nil {
			return err
		}
		user, err := client.UserInfo(context.Background())
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(struct {
			Authenticated bool      `json:"authenticated"`
			User          auth.User `json:"user"`
		}{Authenticated: true, User: user})
	default:
		writeUsage(stderr)
		return errors.New("unknown command")
	}
}

func writeUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n  nala login\n  nala user info\n")
}
