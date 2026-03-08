// tl launches a command in a new terminal window, tab, or pane.
//
// Usage:
//
//	tl [flags] -- <command> [args...]
//
// Examples:
//
//	tl -- vim .
//	tl --in tab -- go test ./...
//	tl --in pane-v --dir ~/project -- make watch
//	tl --in pane-h --name "logs" -- tail -f /var/log/syslog
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pgavlin/tl"
)

const usage = `tl — launch a command in a new terminal window, tab, or pane.

USAGE
  tl [flags] -- <command> [args...]

FLAGS
  --in    <type>      Where to open: window, tab, pane-v, pane-h  (default: tab)
  --dir       <path>      Working directory for the new pane/tab/window
  --name      <title>     Title hint for the window or tab (best-effort)
  --help                  Show this help

EXAMPLES
  tl -- vim .
  tl --in tab -- go test ./...
  tl --in pane-v --dir ~/project -- make watch
`

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, "tl:", err)
		os.Exit(1)
	}
}

func realMain() error {
	var (
		targetStr string
		dir       string
		name      string
		help      bool
	)

	flag.StringVar(&targetStr, "in", "tab", "")
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&name, "name", "", "")
	flag.BoolVar(&help, "help", false, "")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	if help {
		fmt.Fprint(os.Stdout, usage)
		return nil
	}

	args := flag.Args()
	if len(args) == 0 {
		return errors.New("no command supplied\n\n" + strings.TrimRight(usage, "\n"))
	}

	target, err := tl.ParseTarget(targetStr)
	if err != nil {
		return err
	}

	return tl.Launch(args, tl.Options{
		Target: target,
		Dir:    dir,
		Name:   name,
	})
}
