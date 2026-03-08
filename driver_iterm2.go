package tl

import (
	"fmt"
)

// iterm2Driver controls iTerm2 via AppleScript (osascript).
//
// No plugins or extra configuration are required — iTerm2's AppleScript
// dictionary is available by default. The modern Python API is more powerful,
// but requires the API server to be running and an external Python package;
// AppleScript covers all four target types without any setup.
//
// Detection: $TERM_PROGRAM == "iTerm.app" or $ITERM_SESSION_ID is set.
type iterm2Driver struct{}

func (d *iterm2Driver) Name() string { return "iterm2" }

func (d *iterm2Driver) Supports(t Target) bool {
	// iTerm2 supports every target type via AppleScript.
	return true
}

func (d *iterm2Driver) Launch(cmd []string, opts Options) error {
	// We join the command into a single shell string because iTerm2's
	// AppleScript `write text` sends text to the session as if the user typed
	// it — the shell then interprets it.
	cmdStr := appleScriptQuote(shellJoin(cmd))

	cdPrefix := ""
	if opts.Dir != "" {
		cdPrefix = fmt.Sprintf("cd %s && ", appleScriptQuote(opts.Dir))
		// We embed the cd into the command string that write text sends.
		cmdStr = appleScriptQuote("cd " + shellJoin([]string{opts.Dir}) + " && " + shellJoin(cmd))
	}
	_ = cdPrefix // used via cmdStr rebuild above

	var script string
	switch opts.Target {
	case TargetWindow:
		script = fmt.Sprintf(`
tell application "iTerm2"
	set newWindow to (create window with default profile)
	tell current session of newWindow
		write text %s
	end tell
end tell`, cmdStr)

	case TargetTab:
		script = fmt.Sprintf(`
tell application "iTerm2"
	tell current window
		create tab with default profile
		tell current session of current tab
			write text %s
		end tell
	end tell
end tell`, cmdStr)

	case TargetPaneVertical:
		// "split vertically" in iTerm2 means a vertical divider → panes side-by-side.
		script = fmt.Sprintf(`
tell application "iTerm2"
	tell current session of current window
		set s to (split vertically with default profile)
		tell s
			write text %s
		end tell
	end tell
end tell`, cmdStr)

	case TargetPaneHorizontal:
		// "split horizontally" means a horizontal divider → panes stacked.
		script = fmt.Sprintf(`
tell application "iTerm2"
	tell current session of current window
		set s to (split horizontally with default profile)
		tell s
			write text %s
		end tell
	end tell
end tell`, cmdStr)

	default:
		return fmt.Errorf("iterm2: unsupported target %s", opts.Target)
	}

	return run("osascript", "-e", script)
}
