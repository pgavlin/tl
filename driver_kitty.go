package tl

import "fmt"

// kittyDriver controls Kitty via `kitten @` (the remote control interface).
//
// Remote control must be enabled. The simplest way is to add the following to
// kitty.conf:
//
//	allow_remote_control yes
//
// Or launch kitty with: kitty --listen-on unix:/tmp/mykitty
//
// Detection: $KITTY_WINDOW_ID is set in every kitty pane.
type kittyDriver struct{}

func (d *kittyDriver) Name() string { return "kitty" }

func (d *kittyDriver) Supports(t Target) bool {
	// Kitty supports every target type.
	return true
}

func (d *kittyDriver) Launch(cmd []string, opts Options) error {
	// All kitty remote control goes through: kitten @ launch [options] -- cmd
	args := []string{"@", "launch"}

	switch opts.Target {
	case TargetWindow:
		args = append(args, "--type=os-window")
	case TargetTab:
		args = append(args, "--type=tab")
	case TargetPaneVertical:
		// type=window within the current tab; location=vsplit puts it to the right.
		args = append(args, "--type=window", "--location=vsplit")
	case TargetPaneHorizontal:
		args = append(args, "--type=window", "--location=hsplit")
	default:
		return fmt.Errorf("kitty: unsupported target %s", opts.Target)
	}

	if opts.Dir != "" {
		args = append(args, "--cwd="+opts.Dir)
	}
	if opts.Name != "" {
		args = append(args, "--title="+opts.Name)
	}
	// "--" separates kitten flags from the command to run.
	args = append(args, "--")
	args = append(args, cmd...)

	return run("kitten", args...)
}
