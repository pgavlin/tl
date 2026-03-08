// tl launches a command in a new terminal window, tab, or pane.
//
// Usage:
//
//	tl <subcommand> [flags] -- <command> [args...]
//
// Subcommands:
//
//	tab      Open a new tab (default)
//	vsplit   Open a vertical split pane
//	hsplit   Open a horizontal split pane
//	window   Open a new window
//	completion <shell>  Generate shell completions (bash, zsh, fish, pwsh)
//
// Examples:
//
//	tl -- vim .
//	tl tab -- go test ./...
//	tl vsplit --dir ~/project -- make watch
//	tl hsplit --name "logs" -- tail -f /var/log/syslog
//	tl completion bash
package main

import (
	"context"
	"fmt"
	"os"

	cli "github.com/urfave/cli/v3"

	"github.com/pgavlin/tl"
)

func main() {
	if err := root().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "tl:", err)
		os.Exit(1)
	}
}

// launchFlags are shared across all launch subcommands.
func launchFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "dir",
			Usage: "Working directory for the new pane/tab/window",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Title hint for the window or tab (best-effort)",
		},
	}
}

// launchAction returns an Action that launches with the given target.
func launchAction(target tl.Target) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		args := cmd.Args().Slice()
		if len(args) == 0 {
			return fmt.Errorf("no command supplied")
		}
		return tl.Launch(args, tl.Options{
			Target: target,
			Dir:    cmd.String("dir"),
			Name:   cmd.String("name"),
		})
	}
}

func root() *cli.Command {
	return &cli.Command{
		Name:                  "tl",
		Usage:                 "launch a command in a new terminal window, tab, or pane",
		EnableShellCompletion: true,
		DefaultCommand:        "tab",
		Commands: []*cli.Command{
			{
				Name:      "tab",
				Usage:     "Open a new tab (default)",
				Flags:     launchFlags(),
				Action:    launchAction(tl.TargetTab),
				ArgsUsage: "-- <command> [args...]",
			},
			{
				Name:      "vsplit",
				Usage:     "Open a vertical split pane",
				Flags:     launchFlags(),
				Action:    launchAction(tl.TargetPaneVertical),
				ArgsUsage: "-- <command> [args...]",
			},
			{
				Name:      "hsplit",
				Usage:     "Open a horizontal split pane",
				Flags:     launchFlags(),
				Action:    launchAction(tl.TargetPaneHorizontal),
				ArgsUsage: "-- <command> [args...]",
			},
			{
				Name:      "window",
				Usage:     "Open a new window",
				Flags:     launchFlags(),
				Action:    launchAction(tl.TargetWindow),
				ArgsUsage: "-- <command> [args...]",
			},
		},
		// When no recognized subcommand is given, default to tab.
		Action: launchAction(tl.TargetTab),
		Flags:  launchFlags(),
	}
}
