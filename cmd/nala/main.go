package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/azhry/nala-cli/internal/auth"
	"github.com/azhry/nala-cli/internal/platform"
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
	case args[0] == "app":
		return runApp(args[1:], stdout, stderr)
	default:
		writeUsage(stderr)
		return errors.New("unknown command")
	}
}

func writeUsage(writer io.Writer) {
	_, _ = io.WriteString(writer, "Usage:\n  nala login\n  nala user info\n  nala app list [--page N --page-size N]\n  nala app get --id N\n  nala app deploy --id N --source-ref REF --idempotency-key KEY\n  nala app monitor --deployment-id N [--cursor N --follow]\n  nala app delete --id N\n")
}

func runApp(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--help" || args[0] == "-h")) {
		writeUsage(stdout)
		return nil
	}
	switch args[0] {
	case "list":
		return runAppList(args[1:], stdout, stderr)
	case "get":
		return runAppGet(args[1:], stdout, stderr)
	case "deploy":
		return runAppDeploy(args[1:], stdout, stderr)
	case "monitor":
		return runAppMonitor(args[1:], stdout, stderr)
	case "delete":
		return runAppDelete(args[1:], stdout, stderr)
	default:
		writeUsage(stderr)
		return fmt.Errorf("unknown app command %q", args[0])
	}
}

func runAppList(args []string, stdout, stderr io.Writer) error {
	flags := newAppFlagSet("nala app list", stderr, "Usage: nala app list [--page N --page-size N]\n")
	page := flags.Int64("page", 1, "one-based result page")
	pageSize := flags.Int64("page-size", 20, "number of apps per page (1-100)")
	if err := parseAppFlags(flags, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := platform.NewClientFromEnvironment()
	if err != nil {
		return err
	}
	result, err := client.ListApps(context.Background(), *page, *pageSize)
	if err != nil {
		return err
	}
	return encodeOutput(stdout, result)
}

func runAppGet(args []string, stdout, stderr io.Writer) error {
	flags := newAppFlagSet("nala app get", stderr, "Usage: nala app get --id N\n")
	appID := flags.Int64("id", 0, "app identifier")
	if err := parseAppFlags(flags, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := platform.NewClientFromEnvironment()
	if err != nil {
		return err
	}
	result, err := client.GetApp(context.Background(), *appID)
	if err != nil {
		return err
	}
	return encodeOutput(stdout, result)
}

func runAppDeploy(args []string, stdout, stderr io.Writer) error {
	flags := newAppFlagSet("nala app deploy", stderr, "Usage: nala app deploy --id N --source-ref REF --idempotency-key KEY\n")
	appID := flags.Int64("id", 0, "app identifier")
	sourceRef := flags.String("source-ref", "", "source branch, tag, or commit")
	idempotencyKey := flags.String("idempotency-key", "", "unique retry key")
	if err := parseAppFlags(flags, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := platform.NewClientFromEnvironment()
	if err != nil {
		return err
	}
	result, err := client.CreateDeployment(context.Background(), *appID, *sourceRef, *idempotencyKey)
	if err != nil {
		return err
	}
	return encodeOutput(stdout, result)
}

func runAppMonitor(args []string, stdout, stderr io.Writer) error {
	flags := newAppFlagSet("nala app monitor", stderr, "Usage: nala app monitor --deployment-id N [--cursor N --follow]\n")
	deploymentID := flags.Int64("deployment-id", 0, "deployment identifier")
	cursor := flags.Int64("cursor", 0, "return events after this event ID")
	follow := flags.Bool("follow", false, "stream deployment events until the service closes the stream")
	if err := parseAppFlags(flags, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := platform.NewClientFromEnvironment()
	if err != nil {
		return err
	}
	if !*follow {
		result, err := client.GetDeployment(context.Background(), *deploymentID)
		if err != nil {
			return err
		}
		return encodeOutput(stdout, result)
	}
	return client.FollowDeploymentEvents(context.Background(), *deploymentID, *cursor, func(event platform.DeploymentEvent) error {
		return encodeOutput(stdout, event)
	})
}

func runAppDelete(args []string, stdout, stderr io.Writer) error {
	flags := newAppFlagSet("nala app delete", stderr, "Usage: nala app delete --id N\n")
	appID := flags.Int64("id", 0, "app identifier")
	if err := parseAppFlags(flags, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	client, err := platform.NewClientFromEnvironment()
	if err != nil {
		return err
	}
	if err := client.DeleteApp(context.Background(), *appID); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Deleted app %d\n", *appID)
	return nil
}

func newAppFlagSet(name string, stderr io.Writer, usage string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = io.WriteString(stderr, usage)
	}
	return flags
}

func parseAppFlags(flags *flag.FlagSet, args []string) error {
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	return nil
}

func encodeOutput(writer io.Writer, value any) error {
	return json.NewEncoder(writer).Encode(value)
}
