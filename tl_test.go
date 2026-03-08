package tl

import (
	"strings"
	"testing"
)

func TestParseTarget(t *testing.T) {
	cases := []struct {
		input   string
		want    Target
		wantErr bool
	}{
		{"window", TargetWindow, false},
		{"w", TargetWindow, false},
		{"tab", TargetTab, false},
		{"t", TargetTab, false},
		{"pane-v", TargetPaneVertical, false},
		{"vsplit", TargetPaneVertical, false},
		{"pane-vertical", TargetPaneVertical, false},
		{"pane-h", TargetPaneHorizontal, false},
		{"hsplit", TargetPaneHorizontal, false},
		{"pane-horizontal", TargetPaneHorizontal, false},
		// Case-insensitive
		{"TAB", TargetTab, false},
		{"PANE-V", TargetPaneVertical, false},
		// Unknown
		{"split", TargetWindow, true},
		{"", TargetWindow, true},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got, err := ParseTarget(c.input)
			if c.wantErr {
				if err == nil {
					t.Errorf("ParseTarget(%q) = %v, want error", c.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTarget(%q) unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Errorf("ParseTarget(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestTargetString(t *testing.T) {
	cases := map[Target]string{
		TargetWindow:         "window",
		TargetTab:            "tab",
		TargetPaneVertical:   "pane-v",
		TargetPaneHorizontal: "pane-h",
	}
	for target, want := range cases {
		if got := target.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", target, got, want)
		}
	}
}

func TestFallbackChain(t *testing.T) {
	cases := []struct {
		input Target
		want  []Target
	}{
		{
			TargetWindow,
			[]Target{TargetWindow},
		},
		{
			TargetTab,
			[]Target{TargetTab, TargetWindow},
		},
		{
			TargetPaneVertical,
			[]Target{
				TargetPaneVertical,
				TargetPaneHorizontal,
				TargetTab,
				TargetWindow,
			},
		},
		{
			TargetPaneHorizontal,
			[]Target{
				TargetPaneHorizontal,
				TargetPaneVertical,
				TargetTab,
				TargetWindow,
			},
		},
	}

	for _, c := range cases {
		got := fallbackChain(c.input)
		if len(got) != len(c.want) {
			t.Errorf("fallbackChain(%v) len = %d, want %d", c.input, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("fallbackChain(%v)[%d] = %v, want %v", c.input, i, got[i], c.want[i])
			}
		}
	}
}

// ── driver Supports() contracts ───────────────────────────────────────────────

func TestDriverSupports(t *testing.T) {
	allTargets := []Target{
		TargetWindow,
		TargetTab,
		TargetPaneVertical,
		TargetPaneHorizontal,
	}

	// These drivers claim to support everything.
	fullSupport := []struct {
		name   string
		driver terminal
	}{
		{"tmux", &tmuxDriver{}},
		{"wezterm", &wezTermDriver{}},
		{"kitty", &kittyDriver{}},
		{"iterm2", &iterm2Driver{}},
		{"windows-terminal", &windowsTerminalDriver{}},
	}
	for _, tc := range fullSupport {
		for _, tgt := range allTargets {
			if !tc.driver.Supports(tgt) {
				t.Errorf("%s.Supports(%v) = false, want true", tc.name, tgt)
			}
		}
	}

	// Zellij does not support TargetWindow.
	{
		d := &zellijDriver{}
		if d.Supports(TargetWindow) {
			t.Error("zellij.Supports(TargetWindow) = true, want false")
		}
		for _, tgt := range []Target{TargetTab, TargetPaneVertical, TargetPaneHorizontal} {
			if !d.Supports(tgt) {
				t.Errorf("zellij.Supports(%v) = false, want true", tgt)
			}
		}
	}

	// macOS Terminal.app does not support pane splitting.
	{
		d := &macosTerminalDriver{}
		for _, tgt := range []Target{TargetWindow, TargetTab} {
			if !d.Supports(tgt) {
				t.Errorf("macos-terminal.Supports(%v) = false, want true", tgt)
			}
		}
		for _, tgt := range []Target{TargetPaneVertical, TargetPaneHorizontal} {
			if d.Supports(tgt) {
				t.Errorf("macos-terminal.Supports(%v) = true, want false", tgt)
			}
		}
	}
}

// ── Launch fallback ──────────────────────────────────────────────────────────

// stubTerminal is a minimal Terminal that records which target it was called
// with and can be configured to support only a subset of targets.
type stubTerminal struct {
	supported map[Target]bool
	launched  Target
	launchErr error
}

func (s *stubTerminal) Name() string           { return "stub" }
func (s *stubTerminal) Supports(t Target) bool { return s.supported[t] }
func (s *stubTerminal) Launch(_ []string, opts Options) error {
	s.launched = opts.Target
	return s.launchErr
}

func TestLaunchFallback(t *testing.T) {
	cases := []struct {
		name      string
		supported []Target // what the stub supports
		requested Target
		wantUsed  Target
	}{
		{
			name:      "exact match used",
			supported: []Target{TargetPaneVertical, TargetTab, TargetWindow},
			requested: TargetPaneVertical,
			wantUsed:  TargetPaneVertical,
		},
		{
			name:      "pane-v falls back to tab when panes unsupported",
			supported: []Target{TargetTab, TargetWindow},
			requested: TargetPaneVertical,
			wantUsed:  TargetTab,
		},
		{
			name:      "tab falls back to window",
			supported: []Target{TargetWindow},
			requested: TargetTab,
			wantUsed:  TargetWindow,
		},
		{
			name:      "pane-h tries pane-v first",
			supported: []Target{TargetPaneVertical, TargetTab, TargetWindow},
			requested: TargetPaneHorizontal,
			wantUsed:  TargetPaneVertical,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sup := make(map[Target]bool)
			for _, s := range c.supported {
				sup[s] = true
			}
			stub := &stubTerminal{supported: sup}

			err := launchWith(stub, []string{"echo", "hi"}, Options{
				Target: c.requested,
			})
			if err != nil {
				t.Fatalf("launchWith unexpected error: %v", err)
			}
			if stub.launched != c.wantUsed {
				t.Errorf("launched with %v, want %v", stub.launched, c.wantUsed)
			}
		})
	}

	t.Run("error when nothing supported", func(t *testing.T) {
		stub := &stubTerminal{supported: map[Target]bool{}}
		err := launchWith(stub, []string{"echo"}, Options{
			Target: TargetTab,
		})
		if err == nil {
			t.Fatal("expected error when no target is supported")
		}
		if !strings.Contains(err.Error(), "no fallback") {
			t.Errorf("error %q does not mention 'no fallback'", err.Error())
		}
	})
}

// ── Zellij driver Supports ───────────────────────────────────────────────────

func TestZellijDriverSupports(t *testing.T) {
	d := &zellijDriver{}

	supportedTargets := []Target{
		TargetTab,
		TargetPaneVertical,
		TargetPaneHorizontal,
	}
	for _, tgt := range supportedTargets {
		if !d.Supports(tgt) {
			t.Errorf("zellij.Supports(%v) = false, want true", tgt)
		}
	}

	if d.Supports(TargetWindow) {
		t.Error("zellij.Supports(TargetWindow) = true, want false")
	}
}
