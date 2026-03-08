package tl

import "fmt"

// macosTerminalDriver controls macOS Terminal.app via AppleScript (osascript).
//
// Terminal.app does not support pane splitting, so TargetPaneVertical and
// TargetPaneHorizontal are not supported — the fallback chain will try Tab
// and then Window automatically.
//
// Opening a new tab requires sending a Cmd+T keystroke via System Events, which
// means Terminal.app must be in the foreground when the script runs. A short
// delay is inserted to let the tab animate open before the command is sent.
//
// Detection: $TERM_PROGRAM == "Apple_Terminal", or macOS OS-level fallback.
type macosTerminalDriver struct{}

func (d *macosTerminalDriver) Name() string { return "macos-terminal" }

func (d *macosTerminalDriver) Supports(t Target) bool {
	switch t {
	case TargetWindow, TargetTab:
		return true
	case TargetPaneVertical, TargetPaneHorizontal:
		// Terminal.app has no pane splitting.
		return false
	}
	return false
}

func (d *macosTerminalDriver) Launch(cmd []string, opts Options) error {
	cmdStr := appleScriptQuote(shellJoin(cmd))
	if opts.Dir != "" {
		cmdStr = appleScriptQuote("cd " + shellJoin([]string{opts.Dir}) + " && " + shellJoin(cmd))
	}

	var script string
	switch opts.Target {
	case TargetWindow:
		// `do script` with no `in` clause always opens a new window.
		script = fmt.Sprintf(`
tell application "Terminal"
	do script %s
	activate
end tell`, cmdStr)

	case TargetTab:
		// There is no direct AppleScript verb for "new tab" in Terminal.app.
		// The standard workaround is to send Cmd+T via System Events and then
		// run the script in the front window.
		script = fmt.Sprintf(`
tell application "Terminal"
	activate
	tell application "System Events" to keystroke "t" using command down
	delay 0.3
	do script %s in window 1
end tell`, cmdStr)

	default:
		return fmt.Errorf("macos-terminal: unsupported target %s", opts.Target)
	}

	return run("osascript", "-e", script)
}
