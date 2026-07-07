# TODO

Status snapshot for `lazywslcontainer`. Items marked `[x]` are done;
`[ ]` are planned.

## Done

- [x] Scaffold project: `go.mod`, `README.md`, `.gitignore`, `LICENSE`, `CLAUDE.md`
- [x] `wslc` client: types + command wrappers
      (`internal/client/{client,container,image,stats}.go`)
      - version, ping
      - container: list (--all, --format json), logs (--since as unix epoch),
        inspect, stop, start, restart, rm (--force), exec, prune, run
      - image: list (--format json), inspect, rm (--force), prune, build (-t)
      - stats: `--format json` with typed `Stat` struct
- [x] TUI core: bubbletea model/update/view, keybindings (`keys.go`),
      styles (`styles.go`)
- [x] Containers panel: list + logs/stats/inspect tabs + stop/start/restart/remove
- [x] Images panel: list + inspect + rm + run (enter) + build (b)
- [x] Build passes: `go build ./...`, `go vet ./...`, `gofmt -w .`
- [x] Docs: `docs/keybindings.md`, `docs/Config.md`, `README.md`
- [x] Verified JSON shapes against real `wslc 2.9.3.0` output
      (`wslc container list --format json`, `wslc image list --format json`,
      `wslc container inspect`, `wslc image inspect`, `wslc stats --format json`).
      Adjusted `client.Container` and `client.Image` to match:
      - Container: `Id` (not `ID`), `State` is an int enum, `CreatedAt`/
        `StateChangedAt` are unix seconds, `Ports` is an array of
        `{BindingAddress,ContainerPort,HostPort,Protocol}` objects.
      - Image: `Id`, `Size` (int64 bytes), `Created` (int64 unix seconds).
      - Stat: `{ID,Name,CPUPerc,MemUsage,MemPerc,NetIO,BlockIO,PIDs}`.
- [x] Wire `p` (prune) for containers and images with confirmation prompt.
- [x] Wire `b` (build) — inline prompt for `<path> -t <tag>`.
- [x] Wire `enter` on images panel to run a selected image via `client.Run`
      (inline prompt: `[--name N] [-d] [--rm] [-p HOST:CTR] IMAGE [CMD...]`).
- [x] `/` filter mode for side lists (case-insensitive, esc to clear).
- [x] Confirm prompts for destructive actions (`d` remove, `p` prune).
- [x] Mouse support: click selects containers/images and switches tabs;
      scroll wheel scrolls the main panel.
- [x] Scroll in main panel: `pgup`/`pgdown`/`g`/`G` with position indicator.
- [x] Ascii metric graphs: multi-row sparklines (3 rows, `▁▂▃▄▅▆▇█`) with
      history for CPU% and Mem bytes per container. Ring buffer of 60
      samples, sampled every 2s via `wslc stats --format json`. CPU in cyan
      (0-100% scale), Mem in magenta (min-max scale), with min/max/cur labels.
- [x] Inspect pretty-print: JSON flattened (single-element arrays unwrapped),
      keys sorted alphabetically, indented.

## Planned

- [ ] `o` / `e` open config in `$EDITOR` (config tab is a placeholder).
- [ ] Config file loading (`~/.config/lazywslcontainer/config.yml`) —
      keys documented in `docs/Config.md` but parser is not implemented.
- [ ] `commandTemplates` override of `wslc` invocations.
- [ ] `customCommands` per panel (run via `sh -c` / `cmd /c`).
- [ ] Logs follow / tail streaming (currently re-pulls last hour on refresh).
- [ ] Tests: no harness yet. When adding, prefer stubbing `*client.WSLC`
      behind an interface so commands can be faked.
- [ ] Screenshots / demo gif in README.