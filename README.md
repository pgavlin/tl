# tl

A Go package and CLI tool for launching a command in a new terminal window,
tab, or pane. Provides a single, uniform API across major multiplexers and
terminal emulators.

## Supported environments

| Terminal | Window | Tab | Pane (V/H) | Detection |
|---|---|---|---|---|
| **iTerm2** | ✅ | ✅ | ✅ | `$TERM_PROGRAM=iTerm.app` |
| **Kitty** | ✅ | ✅ | ✅ | `$KITTY_WINDOW_ID` |
| **Terminal.app** | ✅ | ✅ | ↓ tab | `$TERM_PROGRAM=Apple_Terminal` |
| **tmux** | ✅ | ✅ | ✅ | `$TMUX` |
| **WezTerm** | ✅ | ✅ | ✅ | `$WEZTERM_PANE` / `$TERM_PROGRAM=WezTerm` |
| **Windows Terminal** | ✅ | ✅ | ✅ | `$WT_SESSION` |
| **Zellij** | ↓ tab | ✅ | ✅ | `$ZELLIJ_SESSION_NAME` |

↓ = falls back gracefully to the next supported target.

## CLI: `tl`

```
tl <subcommand> [flags] -- <command> [args...]

Subcommands:
  tab      Open a new tab (default)
  vsplit   Open a vertical split pane
  hsplit   Open a horizontal split pane
  window   Open a new window

Flags:
  --dir    working directory
  --name   title for the new window/tab/pane
```

### Examples

```bash
# Open vim in a new tab (auto-detects terminal)
tl -- vim .

# Open test output in a vertical split
tl vsplit -- go test -v ./...

# Tail a log in a horizontal pane in a specific directory
tl hsplit --dir /var/log --name "app logs" -- tail -f app.log

# Generate shell completions
tl completion bash
```

### Install

```bash
go install github.com/pgavlin/tl/cmd/tl@latest
```

## Package API

```go
import "github.com/pgavlin/tl"

// Auto-detect terminal, open a new tab.
err := tl.Launch([]string{"vim", "."}, tl.Options{
    Target: tl.TargetTab,
})

// Open a vertical split with a working directory and title.
err = tl.Launch([]string{"htop"}, tl.Options{
    Target: tl.TargetPaneVertical,
    Dir:    "/var/log",
    Name:   "system monitor",
})
```

### Options

```go
type Options struct {
    Target Target // TargetWindow | TargetTab | TargetPaneVertical | TargetPaneHorizontal
    Dir    string // working directory (empty = inherit)
    Name   string // title hint (best-effort; not all terminals support it)
}
```

## Fallback behaviour

When the requested target is not supported by the detected terminal, the library
automatically tries the next best option rather than returning an error:

```
TargetPaneVertical   → TargetPaneHorizontal → TargetTab → TargetWindow
TargetPaneHorizontal → TargetPaneVertical   → TargetTab → TargetWindow
TargetTab            → TargetWindow
TargetWindow         → (error if nothing works)
```

## Terminal-specific notes

### Kitty

Requires `allow_remote_control yes` in `kitty.conf`.

### Zellij

Tab-level launches write a temporary KDL layout file consumed by
`zellij action new-tab --layout`. The file is deleted after the command runs.
Zellij has no OS window concept; `TargetWindow` falls back to `TargetTab`.

### macOS Terminal.app

Tab creation uses a `System Events` keystroke (Cmd+T) because Terminal.app has
no AppleScript verb for "new tab". Terminal.app must be in the foreground when
the command runs. Pane splitting is not supported.

### iTerm2

Uses AppleScript (`osascript`). No extra configuration needed.
For richer automation, consider using the iTerm2 Python API directly.

## License

MIT
