package tl

import (
	"os"
	"strings"
)

// tmuxDriver controls tmux via the tmux CLI.
//
// In tmux, a "window" is what most emulators call a "tab" — a named entry in
// the status bar. Panes are splits inside a window. Both TargetWindow and
// TargetTab map to new-window; the distinction is meaningful for other
// emulators but semantically equivalent here.
//
// Detection: $TMUX is set when running inside a tmux session.
type tmuxDriver struct{}

func (d *tmuxDriver) Name() string { return "tmux" }

func (d *tmuxDriver) Supports(t Target) bool {
	// tmux supports every target type.
	return true
}

func (d *tmuxDriver) Launch(cmd []string, opts Options) error {
	switch opts.Target {
	case TargetWindow, TargetTab:
		return d.newWindow(cmd, opts)
	case TargetPaneVertical:
		return d.splitWindow(cmd, opts, "-h") // -h = split left/right (vertical divider)
	case TargetPaneHorizontal:
		return d.splitWindow(cmd, opts, "-v") // -v = split top/bottom (horizontal divider)
	}
	return nil
}

func (d *tmuxDriver) newWindow(cmd []string, opts Options) error {
	args := []string{"new-window"}
	if opts.Dir != "" {
		args = append(args, "-c", opts.Dir)
	}
	if opts.Name != "" {
		args = append(args, "-n", opts.Name)
	}
	// tmux new-window takes the command as a single shell string at the end.
	if len(cmd) > 0 {
		args = append(args, shellJoin(cmd))
	}
	return run("tmux", args...)
}

func (d *tmuxDriver) splitWindow(cmd []string, opts Options, splitFlag string) error {
	args := []string{"split-window", splitFlag}

	cwd := opts.Dir
	if cwd == "" {
		// Inherit the caller's working directory rather than the shell default.
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}
	if cwd != "" {
		args = append(args, "-c", cwd)
	}

	if len(cmd) > 0 {
		args = append(args, strings.Join(cmd, " "))
	}
	return run("tmux", args...)
}
