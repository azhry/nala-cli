package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/azhry/nala-cli/internal/auth"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "nala: %s\n", err)
		os.Exit(1)
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
