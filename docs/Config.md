# Config

`lazywslcontainer` reads its config from:

- `%USERPROFILE%\.config\lazywslcontainer\config.yml` (Windows)
- `~/.config/lazywslcontainer/config.yml` (Linux/macOS if/when ported)

The file is optional — sensible defaults are used if absent.

> Config loading is planned; the keys below are the target surface and are
> partially hard-coded today. Patches welcome.

## Example

```yaml
gui:
  wrapMainPanel: false
  theme: dark          # dark | light | custom
  showAllContainers: true

commandTemplates:
  containerLogs: "wslc container logs {{.ID}} --since {{.SinceUnix}}"
  containerStop: "wslc container stop {{.ID}}"
  containerRemove: "wslc container rm {{.ID}} --force"
  imageRemove: "wslc image rm {{.ID}} --force"
  stats: "wslc stats --format json"

customCommands:
  containers:
    - name: shell
      command: "wslc exec {{.Name}} bash"
      description: "attach a shell"
```

## Keys

### `gui`

| Key                | Default | Description                                  |
| ------------------ | ------- | -------------------------------------------- |
| `wrapMainPanel`    | `false` | Wrap long lines in the main panel (CPU cost) |
| `theme`            | `dark`  | Color theme                                   |
| `showAllContainers`| `true`  | Pass `--all` to `wslc container list`         |

### `commandTemplates`

Each value is a Go text/template executed against the selected item. Lets
you override the exact `wslc` invocation — useful while `wslc`'s flags are
still stabilising. See the [WSL containers docs](https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-containers).

### `customCommands`

Per-panel custom commands surfaced in the `config` tab. `command` is run via
`sh -c` (Linux) / `cmd /c` (Windows); `{{.Name}}` / `{{.ID}}` interpolate the
selected item.