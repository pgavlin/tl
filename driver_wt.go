package tl

import "fmt"

// windowsTerminalDriver controls Windows Terminal via the wt.exe CLI.
//
// wt.exe is available in PATH when "App execution aliases" are enabled in
// Windows Settings (the default). Multiple subcommands can be chained with
// semicolons on the command line.
//
// The --window flag selects the target window:
//   - "new" / -1 → always open a fresh OS window
//   - "0" / "last" → the most recently used window
//
// Detection: $WT_SESSION is set in every Windows Terminal pane.
type windowsTerminalDriver struct{}

func (d *windowsTerminalDriver) Name() string { return "windows-terminal" }

func (d *windowsTerminalDriver) Supports(t Target) bool {
	// Windows Terminal supports every target type.
	return true
}

func (d *windowsTerminalDriver) Launch(cmd []string, opts Options) error {
	switch opts.Target {
	case TargetWindow:
		// -w new always creates a fresh OS window.
		args := []string{"-w", "new", "new-tab"}
		args = wtFlags(args, opts)
		args = append(args, cmd...)
		return run("wt.exe", args...)

	case TargetTab:
		// -w 0 targets the most recently used existing window.
		args := []string{"-w", "0", "new-tab"}
		args = wtFlags(args, opts)
		args = append(args, cmd...)
		return run("wt.exe", args...)

	case TargetPaneVertical:
		// -V splits the focused pane into two side-by-side panes.
		args := []string{"-w", "0", "split-pane", "-V"}
		args = wtFlags(args, opts)
		args = append(args, cmd...)
		return run("wt.exe", args...)

	case TargetPaneHorizontal:
		// -H splits the focused pane into two stacked panes.
		args := []string{"-w", "0", "split-pane", "-H"}
		args = wtFlags(args, opts)
		args = append(args, cmd...)
		return run("wt.exe", args...)
	}

	return fmt.Errorf("windows-terminal: unsupported target %s", opts.Target)
}

// wtFlags appends common optional flags (directory, title) to the wt argument
// list and returns the updated slice.
func wtFlags(args []string, opts Options) []string {
	if opts.Dir != "" {
		args = append(args, "-d", opts.Dir)
	}
	if opts.Name != "" {
		args = append(args, "--title", opts.Name)
	}
	return args
}
