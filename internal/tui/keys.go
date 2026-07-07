package tui

// Keybindings centralises the key strings used across the TUI.
// Mirrors lazydocker's defaults where they map cleanly onto wslc.
type Keys struct {
	Quit      []string
	Up        []string
	Down      []string
	Left      []string
	Right     []string
	NextPanel []string
	PrevPanel []string

	Stop       []string
	Start      []string
	Restart    []string
	Remove     []string
	Prune      []string
	Build      []string
	Run        []string
	Refresh    []string
	Tab        []string
	Filter     []string
	OpenConfig []string
	ScrollUp   []string
	ScrollDown []string
	ScrollTop  []string
	ScrollBot  []string
}

func DefaultKeys() Keys {
	return Keys{
		Quit:       []string{"q", "ctrl+c", "esc"},
		Up:         []string{"up", "k"},
		Down:       []string{"down", "j"},
		Left:       []string{"left", "h"},
		Right:      []string{"right", "l"},
		NextPanel:  []string{"tab"},
		PrevPanel:  []string{"shift+tab"},
		Stop:       []string{"s"},
		Start:      []string{"S"},
		Restart:    []string{"r"},
		Remove:     []string{"d"},
		Prune:      []string{"p"},
		Build:      []string{"b"},
		Run:        []string{"enter"},
		Refresh:    []string{"R"},
		Tab:        []string{"[", "]"},
		Filter:     []string{"/"},
		OpenConfig: []string{"o", "e"},
		ScrollUp:   []string{"pgup", "K"},
		ScrollDown: []string{"pgdown", "J"},
		ScrollTop:  []string{"g"},
		ScrollBot:  []string{"G"},
	}
}

// Panel identifies which side/main panel has focus.
type Panel int

const (
	PanelContainers Panel = iota
	PanelImages
	PanelMain
)

// MainTab identifies which tab within the main panel is shown.
type MainTab int

const (
	TabLogs MainTab = iota
	TabStats
	TabInspect
	TabConfig
)

func (t MainTab) String() string {
	switch t {
	case TabLogs:
		return "logs"
	case TabStats:
		return "stats"
	case TabInspect:
		return "inspect"
	case TabConfig:
		return "config"
	}
	return "?"
}
