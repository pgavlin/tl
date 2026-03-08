package tl

import (
	"fmt"
	"os"
	"strings"
)

// zellijDriver controls Zellij via `zellij` CLI and `zellij action`.
//
// Zellij has sessions and tabs but no OS-level windows, so TargetWindow falls
// back to TargetTab via the standard fallback chain.
//
// Running a command in a new pane uses `zellij run`, which is first-class.
// Running a command in a new tab requires writing a temporary KDL layout file,
// since `zellij action new-tab` alone does not accept a commandline.
//
// Detection: $ZELLIJ_SESSION_NAME or $ZELLIJ is set inside a session.
type zellijDriver struct{}

func (d *zellijDriver) Name() string { return "zellij" }

func (d *zellijDriver) Supports(t Target) bool {
	switch t {
	case TargetTab, TargetPaneVertical, TargetPaneHorizontal:
		return true
	case TargetWindow:
		// Zellij has no OS window concept — fallback to tab.
		return false
	}
	return false
}

func (d *zellijDriver) Launch(cmd []string, opts Options) error {
	switch opts.Target {
	case TargetTab:
		return d.newTab(cmd, opts)
	case TargetPaneVertical:
		return d.runPane(cmd, opts, "right")
	case TargetPaneHorizontal:
		return d.runPane(cmd, opts, "down")
	}
	return fmt.Errorf("zellij: unsupported target %s", opts.Target)
}

// runPane uses `zellij run` to open a command pane in the given direction.
func (d *zellijDriver) runPane(cmd []string, opts Options, direction string) error {
	args := []string{"run", "--direction", direction}
	if opts.Dir != "" {
		args = append(args, "--cwd", opts.Dir)
	}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	args = append(args, "--")
	args = append(args, cmd...)
	return run("zellij", args...)
}

// newTab opens a new Zellij tab running cmd by writing a temporary KDL layout
// file and passing it to `zellij action new-tab --layout`.
func (d *zellijDriver) newTab(cmd []string, opts Options) error {
	cwd := opts.Dir
	if cwd == "" {
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}

	kdl, err := buildZellijLayout(cmd, cwd, opts.Name)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp("", "tl-*.kdl")
	if err != nil {
		return fmt.Errorf("zellij: create temp layout: %w", err)
	}
	defer os.Remove(f.Name()) // clean up after zellij has read it

	if _, err := f.WriteString(kdl); err != nil {
		f.Close()
		return fmt.Errorf("zellij: write temp layout: %w", err)
	}
	f.Close()

	return run("zellij", "action", "new-tab", "--layout", f.Name())
}

// buildZellijLayout produces a minimal KDL layout that opens one pane running
// cmd. The layout is written to a temp file consumed by `zellij action new-tab`.
//
// KDL layout structure (Zellij ≥ 0.32):
//
//	layout {
//	    tab name="title" {
//	        pane cwd="/path" {
//	            command "prog"
//	            args "arg1" "arg2"
//	        }
//	    }
//	}
func buildZellijLayout(cmd []string, cwd, name string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("zellij: command must not be empty")
	}

	tabName := name
	if tabName == "" {
		tabName = cmd[0]
	}

	var sb strings.Builder
	sb.WriteString("layout {\n")
	fmt.Fprintf(&sb, "    tab name=%s {\n", kdlQuote(tabName))
	if cwd != "" {
		fmt.Fprintf(&sb, "        pane cwd=%s {\n", kdlQuote(cwd))
	} else {
		sb.WriteString("        pane {\n")
	}
	fmt.Fprintf(&sb, "            command %s\n", kdlQuote(cmd[0]))
	if len(cmd) > 1 {
		sb.WriteString("            args")
		for _, arg := range cmd[1:] {
			fmt.Fprintf(&sb, " %s", kdlQuote(arg))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

// kdlQuote wraps s in KDL double-quoted string literals.
func kdlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
