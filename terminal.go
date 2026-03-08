// Package tl provides a unified API for opening a command in a new
// terminal window, tab, or pane across multiple terminal emulators and
// multiplexers.
//
// Quick start:
//
//	err := tl.Launch([]string{"vim", "."}, tl.Options{
//	    Target: tl.TargetTab,
//	})
//
// The terminal is auto-detected from environment variables. If the requested
// target type (e.g. a pane split) is not supported, the library falls back
// gracefully to the next best option.
package tl

import (
	"fmt"
	"os/exec"
	"strings"
)

// Target describes where the new command should be opened.
type Target int

const (
	// TargetWindow opens a new top-level OS window.
	TargetWindow Target = iota
	// TargetTab opens a new tab inside the current window.
	TargetTab
	// TargetPaneVertical opens a vertical split (new pane to the right).
	TargetPaneVertical
	// TargetPaneHorizontal opens a horizontal split (new pane below).
	TargetPaneHorizontal
)

func (t Target) String() string {
	switch t {
	case TargetWindow:
		return "window"
	case TargetTab:
		return "tab"
	case TargetPaneVertical:
		return "pane-v"
	case TargetPaneHorizontal:
		return "pane-h"
	}
	return "unknown"
}

// ParseTarget converts a string to a Target. Accepted values:
// "window" / "w", "tab" / "t", "pane-v" / "vsplit", "pane-h" / "hsplit".
func ParseTarget(s string) (Target, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "window", "w":
		return TargetWindow, nil
	case "tab", "t":
		return TargetTab, nil
	case "pane-v", "pane-vertical", "vsplit":
		return TargetPaneVertical, nil
	case "pane-h", "pane-horizontal", "hsplit":
		return TargetPaneHorizontal, nil
	}
	return TargetWindow, fmt.Errorf(
		"unknown target %q: accepted values are window, tab, pane-v, pane-h", s)
}

// Options configures how a command is launched.
type Options struct {
	// Target is where to open the command. Defaults to TargetTab.
	Target Target

	// Dir is the working directory for the new pane/tab/window.
	// An empty string inherits the terminal's default (usually the current dir).
	Dir string

	// Name is a title hint for the new window, tab, or pane.
	// Not all terminals honour this; it is applied on a best-effort basis.
	Name string
}

// terminal is the interface that every driver must implement.
type terminal interface {
	// Name returns a stable, human-readable identifier such as "tmux" or "kitty".
	Name() string
	// Supports reports whether the driver is able to open the given target type.
	Supports(t Target) bool
	// Launch opens cmd in the requested target. opts.Target is guaranteed to be
	// a value for which Supports returned true.
	Launch(cmd []string, opts Options) error
}

// Launch auto-detects the running terminal and opens cmd with automatic
// fallback if the exact target type is unsupported.
func Launch(cmd []string, opts Options) error {
	term, err := detect()
	if err != nil {
		return fmt.Errorf("tl: %w", err)
	}
	return launchWith(term, cmd, opts)
}

func launchWith(term terminal, cmd []string, opts Options) error {
	for _, t := range fallbackChain(opts.Target) {
		if term.Supports(t) {
			o := opts
			o.Target = t
			if err := term.Launch(cmd, o); err != nil {
				return fmt.Errorf("tl (%s, %s): %w", term.Name(), t, err)
			}
			return nil
		}
	}
	return fmt.Errorf("tl: %s supports no fallback target for %s",
		term.Name(), opts.Target)
}

func fallbackChain(t Target) []Target {
	switch t {
	case TargetPaneVertical:
		return []Target{TargetPaneVertical, TargetPaneHorizontal, TargetTab, TargetWindow}
	case TargetPaneHorizontal:
		return []Target{TargetPaneHorizontal, TargetPaneVertical, TargetTab, TargetWindow}
	case TargetTab:
		return []Target{TargetTab, TargetWindow}
	default:
		return []Target{TargetWindow}
	}
}

// ─── shared helpers ──────────────────────────────────────────────────────────

// run executes the named binary with args, returning a descriptive error on
// failure that includes the combined stdout+stderr output.
func run(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("%s: %w", name, err)
		}
		return fmt.Errorf("%s: %w\n%s", name, err, msg)
	}
	return nil
}

// appleScriptQuote wraps s in AppleScript double-quote literals, escaping
// backslashes and double quotes so the result is safe to embed in a script.
func appleScriptQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// shellJoin concatenates words into a single shell command string.
// Words that contain spaces or special characters are single-quoted.
func shellJoin(words []string) string {
	parts := make([]string, len(words))
	for i, w := range words {
		if strings.ContainsAny(w, " \t\n\"'\\|&;()<>{}$`!#~") {
			w = "'" + strings.ReplaceAll(w, "'", `'\''`) + "'"
		}
		parts[i] = w
	}
	return strings.Join(parts, " ")
}
