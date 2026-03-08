package tl

// wezTermDriver controls WezTerm via `wezterm cli`.
//
// The CLI is available from within a running WezTerm window without any extra
// configuration — WezTerm exposes a socket automatically.
//
// Detection: $WEZTERM_PANE is set in every pane, or $TERM_PROGRAM == "WezTerm".
type wezTermDriver struct{}

func (d *wezTermDriver) Name() string { return "wezterm" }

func (d *wezTermDriver) Supports(t Target) bool {
	// WezTerm supports every target type.
	return true
}

func (d *wezTermDriver) Launch(cmd []string, opts Options) error {
	switch opts.Target {
	case TargetWindow:
		args := []string{"cli", "spawn", "--new-window"}
		args = weztermCWD(args, opts.Dir)
		args = append(args, "--")
		args = append(args, cmd...)
		return run("wezterm", args...)

	case TargetTab:
		args := []string{"cli", "spawn"}
		args = weztermCWD(args, opts.Dir)
		args = append(args, "--")
		args = append(args, cmd...)
		return run("wezterm", args...)

	case TargetPaneVertical:
		// --right splits side-by-side (vertical divider).
		args := []string{"cli", "split-pane", "--right"}
		args = weztermCWD(args, opts.Dir)
		args = append(args, "--")
		args = append(args, cmd...)
		return run("wezterm", args...)

	case TargetPaneHorizontal:
		// --bottom splits top-and-bottom (horizontal divider).
		args := []string{"cli", "split-pane", "--bottom"}
		args = weztermCWD(args, opts.Dir)
		args = append(args, "--")
		args = append(args, cmd...)
		return run("wezterm", args...)
	}
	return nil
}

func weztermCWD(args []string, dir string) []string {
	if dir != "" {
		return append(args, "--cwd", dir)
	}
	return args
}
