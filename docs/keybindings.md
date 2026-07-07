# Keybindings

Default keys. Case-sensitive where two bindings share a letter (e.g. `s` stop
vs `S` start).

## Global

| Key     | Action                          |
| ------- | ------------------------------- |
| `q`     | Quit (also `ctrl+c`, `esc`)     |
| `tab`   | Next panel                      |
| `S-tab` | Previous panel                  |
| `[` / `]` | Previous / next main tab       |
| `b`     | Build an image (inline prompt)  |
| `R`     | Force refresh from `wslc`       |
| `pgup`/`K` | Scroll main panel up (10 lines) |
| `pgdown`/`J` | Scroll main panel down |
| `g` / `G` | Scroll to top / bottom of main panel |
| `/`     | Filter side list (inline prompt, esc to clear) |
| `o`/`e` | Open config in `$EDITOR` (planned) |

When an inline prompt is open: type to edit, `enter` to confirm, `esc` to
cancel.

## Containers panel

| Key   | Action                |
| ----- | --------------------- |
| `j` / `k` | Move selection down / up |
| `s`   | Stop selected container |
| `S`   | Start selected container |
| `r`   | Restart selected container |
| `d`   | Remove (force) selected container |
| `p`   | Prune all stopped containers |

## Images panel

| Key   | Action                |
| ----- | --------------------- |
| `j` / `k` | Move selection down / up |
| `enter` | Run selected image (inline prompt: `[--name N] [-d] [--rm] [-p HOST:CTR] IMAGE [CMD...]`) |
| `d`   | Remove (force) selected image |
| `p`   | Prune unused images |

## Main tabs

Tabs cycle through `logs`, `stats`, `inspect`, `config` via `[` / `]`.

- **logs** — `wslc container logs <id>` (last hour, follows on refresh)
- **stats** — `wslc stats --no-stream`
- **inspect** — `wslc container inspect <id>` (or `wslc image inspect <id>` on the images panel)
- **config** — reserved for editing `lazywslcontainer`'s config (planned)

## Mouse

| Action | Effect |
| ----- | ----- |
| Click on a side list row | Select that container/image and focus the panel |
| Click on a tab name | Switch to that tab |
| Scroll wheel | Scroll the main panel (inspect/logs) |

## Confirm prompts

Destructive actions (`d` remove, `p` prune) show a `[y/n]` prompt.
`y`/`enter` confirms, `n`/`esc` cancels.