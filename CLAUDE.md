# CLAUDE.md

Project guidance for AI agents working on `lazywslcontainer`.

## What this is

A TUI for `wslc.exe`, the WSL container CLI documented at
https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-containers.
Inspired by [lazydocker](https://github.com/jesseduffield/lazydocker) but
unaffiliated.

## Tech stack

- Go 1.24
- [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI loop
- [lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- No CGO, no external runtime deps beyond `wslc.exe` on PATH

## Layout

```
main.go                       # entrypoint, ping check, launch tea.Program
internal/client/*.go         # thin wrappers over `wslc` subcommands
internal/tui/                 # bubbletea model/update/view, panels, tabs
docs/                         # keybindings + config docs
```

## Build / run

```powershell
go build ./...
go build -o lazywslcontainer.exe .
./lazywslcontainer.exe
```

There is no test harness yet. When adding client code, prefer thin wrappers
that are easy to stub via an interface if/when tests arrive.

## wslc surface (as documented)

```
wslc version
wslc run [--rm] [-d] [-it] [-p HOST:CTR] [--name N] IMAGE [CMD...]
wslc exec CONTAINER CMD...
wslc build -t TAG CONTEXT
wslc container list [--all] [--format json]
wslc container logs CONTAINER [--since NDs]
wslc container inspect CONTAINER
wslc container stop|start|restart CONTAINER
wslc container rm CONTAINER [--force]
wslc container prune
wslc image list [--format json]
wslc image inspect IMAGE
wslc image rm IMAGE [--force]
wslc image prune
wslc stats [--no-stream]
```

If `wslc`'s flags change, update `internal/client/` first — the TUI should
never shell out directly.

## Conventions

- No comments unless asked.
- Follow existing patterns in the package you're editing.
- Style via `internal/tui/styles.go` only; don't inline colors.
- Keep `wslc` invocations in `internal/client/` so the TUI stays mockable.
- Don't commit `lazywslcontainer.exe`.