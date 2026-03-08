package tl

import (
	"fmt"
	"os"
)

// detect inspects the process environment and returns the best-matching
// terminal driver. Multiplexers (tmux, Zellij) are checked first because they
// run nested inside other emulators.
//
// Detection heuristics, in priority order:
//
//  1. $TMUX set                     → tmux
//  2. $ZELLIJ_SESSION_NAME set      → zellij
//  3. $KITTY_WINDOW_ID set          → kitty
//  4. $WEZTERM_PANE set             → wezterm
//  5. $TERM_PROGRAM == "iTerm.app"  → iterm2
//  6. $TERM_PROGRAM == "WezTerm"    → wezterm
//  7. $TERM_PROGRAM == "kitty"      → kitty
//  8. $TERM_PROGRAM == "Apple_Terminal" → macos-terminal
//  9. $WT_SESSION set               → windows-terminal
//  10. darwin OS (fallback)         → macos-terminal
func detect() (terminal, error) {
	if os.Getenv("TMUX") != "" {
		return &tmuxDriver{}, nil
	}
	if os.Getenv("ZELLIJ_SESSION_NAME") != "" || os.Getenv("ZELLIJ") != "" {
		return &zellijDriver{}, nil
	}

	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return &kittyDriver{}, nil
	}
	if os.Getenv("WEZTERM_PANE") != "" {
		return &wezTermDriver{}, nil
	}

	switch tp := os.Getenv("TERM_PROGRAM"); tp {
	case "iTerm.app":
		return &iterm2Driver{}, nil
	case "WezTerm":
		return &wezTermDriver{}, nil
	case "kitty":
		return &kittyDriver{}, nil
	case "Apple_Terminal":
		return &macosTerminalDriver{}, nil
	}

	if os.Getenv("WT_SESSION") != "" {
		return &windowsTerminalDriver{}, nil
	}

	return nil, fmt.Errorf("could not detect terminal from environment; set TERM_PROGRAM")
}
