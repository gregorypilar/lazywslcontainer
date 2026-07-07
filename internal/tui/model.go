package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gregorypilar/lazywslcontainer/internal/client"
)

// Model is the root bubbletea model for lazywslcontainer.
type Model struct {
	keys   Keys
	styles Styles

	client *client.WSLC

	width, height int

	// panels
	panel Panel

	// side lists
	containers []client.Container
	images     []client.Image
	curCont    int
	curImg     int

	// main panel
	tab          MainTab
	logs         []string
	statsHistory map[string]*containerStats
	inspect      string
	scroll       int
	err          error

	// status / help
	status string

	lastRefresh time.Time
	ready       bool

	// inline input mode (run / build prompts)
	inputMode  inputMode
	inputLabel string
	inputBuf   string
	inputSeed  string

	// confirm mode (destructive actions)
	confirmAction func() tea.Cmd
	confirmLabel  string

	// filter
	filterCont string
	filterImg  string
}

type inputMode int

const (
	inputNone inputMode = iota
	inputRun
	inputBuild
	inputFilter
)

// messages
type tickMsg time.Time
type refreshMsg struct {
	containers []client.Container
	images     []client.Image
	err        error
}
type logsMsg struct {
	lines []string
	err   error
}
type statsMsg struct {
	stats []client.Stat
	err   error
}
type inspectMsg struct {
	text string
	err  error
}
type actionDoneMsg struct {
	action, target string
	err            error
}
type runResultMsg struct {
	out []byte
	err error
}

// New returns the initial Model.
func New(c *client.WSLC) Model {
	return Model{
		keys:   DefaultKeys(),
		styles: DefaultStyles(),
		client: c,
		panel:  PanelContainers,
		tab:    TabLogs,
		status: "starting up…",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(refresh(m.client), tick(), sampleStats(m.client))
}

// ---- Update -------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = int(msg.Width), int(msg.Height)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tickMsg:
		return m, tea.Batch(refresh(m.client), tick(), sampleStats(m.client))

	case refreshMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("refresh failed: %v", msg.err)
		} else {
			m.err = nil
			m.containers = msg.containers
			m.images = msg.images
			m.lastRefresh = time.Now()
			m.status = fmt.Sprintf("containers: %d, images: %d — %s",
				len(msg.containers), len(msg.images), m.lastRefresh.Format("15:04:05"))
		}
		return m, nil

	case logsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.logs = []string{fmt.Sprintf("logs error: %v", msg.err)}
		} else {
			m.logs = strings.Split(strings.TrimSpace(string(strings.Join(msg.lines, "\n"))), "\n")
		}
		m.scroll = 0
		return m, nil

	case statsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("stats error: %v", msg.err)
		} else {
			m.err = nil
			m.ingestStats(msg.stats)
		}
		return m, nil

	case inspectMsg:
		if msg.err != nil {
			m.err = msg.err
			m.inspect = fmt.Sprintf("inspect error: %v", msg.err)
		} else {
			m.inspect = prettyInspect(msg.text)
		}
		m.scroll = 0
		return m, nil

	case actionDoneMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("%s %s failed: %v", msg.action, msg.target, msg.err)
		} else {
			m.status = fmt.Sprintf("%s %s ok", msg.action, msg.target)
		}
		return m, tea.Batch(refresh(m.client))

	case runResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("run failed: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("run ok: %s", strings.TrimSpace(string(msg.out)))
		}
		return m, tea.Batch(refresh(m.client))
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.confirmAction != nil {
		return m.handleConfirmKey(msg)
	}
	if m.inputMode != inputNone {
		return m.handleInputKey(msg)
	}
	if contains(m.keys.Quit, msg.String()) {
		return m, tea.Quit
	}
	if contains(m.keys.Refresh, msg.String()) {
		return m, refresh(m.client)
	}
	if contains(m.keys.NextPanel, msg.String()) {
		m.panel = (m.panel + 1) % 3
		m.scroll = 0
		return m, m.loadCurrentTab()
	}
	if contains(m.keys.PrevPanel, msg.String()) {
		m.panel = (m.panel + 2) % 3
		m.scroll = 0
		return m, m.loadCurrentTab()
	}
	if contains(m.keys.Tab, "]") && msg.String() == "]" {
		m.tab = (m.tab + 1) % 4
		m.scroll = 0
		return m, m.loadCurrentTab()
	}
	if contains(m.keys.Tab, "[") && msg.String() == "[" {
		m.tab = (m.tab + 3) % 4
		m.scroll = 0
		return m, m.loadCurrentTab()
	}
	if contains(m.keys.Build, msg.String()) {
		m.enterBuildPrompt()
		return m, nil
	}
	if contains(m.keys.Filter, msg.String()) {
		m.enterFilterPrompt()
		return m, nil
	}

	if contains(m.keys.ScrollUp, msg.String()) {
		m.scrollUp(10)
		return m, nil
	}
	if contains(m.keys.ScrollDown, msg.String()) {
		m.scrollDown(10)
		return m, nil
	}
	if contains(m.keys.ScrollTop, msg.String()) {
		m.scroll = 0
		return m, nil
	}
	if contains(m.keys.ScrollBot, msg.String()) {
		m.scroll = -1
		return m, nil
	}

	switch m.panel {
	case PanelContainers:
		return m.handleContainerKey(msg)
	case PanelImages:
		return m.handleImageKey(msg)
	}
	return m, nil
}

func (m Model) handleContainerKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	s := msg.String()
	filt := m.filteredContainers()
	switch {
	case contains(m.keys.Up, s):
		if m.curCont > 0 {
			m.curCont--
		}
		m.scroll = 0
		return m, m.loadCurrentTab()
	case contains(m.keys.Down, s):
		if m.curCont < len(filt)-1 {
			m.curCont++
		}
		m.scroll = 0
		return m, m.loadCurrentTab()
	case contains(m.keys.Stop, s):
		return m, runAction(m.client, "stop", m.curContainerID(), func(c *client.WSLC, ctx context.Context, id string) error {
			return c.ContainerStop(ctx, id)
		})
	case contains(m.keys.Start, s):
		return m, runAction(m.client, "start", m.curContainerID(), func(c *client.WSLC, ctx context.Context, id string) error {
			return c.ContainerStart(ctx, id)
		})
	case contains(m.keys.Restart, s):
		return m, runAction(m.client, "restart", m.curContainerID(), func(c *client.WSLC, ctx context.Context, id string) error {
			return c.ContainerRestart(ctx, id)
		})
	case contains(m.keys.Remove, s):
		id := m.curContainerID()
		return m.withConfirm(fmt.Sprintf("remove container %s?", id), func() tea.Cmd {
			return runAction(m.client, "rm", id, func(c *client.WSLC, ctx context.Context, id string) error {
				return c.ContainerRemove(ctx, id, true)
			})
		}), nil
	case contains(m.keys.Prune, s):
		return m.withConfirm("prune all stopped containers?", func() tea.Cmd {
			return pruneAction(m.client, "container prune", func(c *client.WSLC, ctx context.Context) ([]byte, error) {
				return c.ContainerPrune(ctx)
			})
		}), nil
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	x, y := int(msg.X), int(msg.Y)

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scrollUp(3)
		return m, nil
	case tea.MouseButtonWheelDown:
		m.scrollDown(3)
		return m, nil
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		return m.handleClick(x, y)
	}
	return m, nil
}

func (m Model) handleClick(x, y int) (Model, tea.Cmd) {
	sideW := m.width / 4
	if sideW < 24 {
		sideW = 24
	}
	if y == 0 {
		return m, nil
	}
	if x < sideW {
		contW := sideW / 2
		if x < contW {
			m.panel = PanelContainers
			filt := m.filteredContainers()
			idx := y - 1
			if idx >= 0 && idx < len(filt) {
				m.curCont = idx
			}
		} else {
			m.panel = PanelImages
			filt := m.filteredImages()
			idx := y - 1
			if idx >= 0 && idx < len(filt) {
				m.curImg = idx
			}
		}
		m.scroll = 0
		return m, m.loadCurrentTab()
	}
	mainX := sideW + 3
	if x >= mainX {
		m.panel = PanelMain
		tabAreaW := m.width - mainX
		tabNames := []string{"logs", "stats", "inspect", "config"}
		tabX := 0
		for i, name := range tabNames {
			w := len(name) + 1
			if x-mainX >= tabX && x-mainX < tabX+w {
				m.tab = MainTab(i)
				m.scroll = 0
				return m, m.loadCurrentTab()
			}
			tabX += w
		}
		_ = tabAreaW
	}
	return m, nil
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	s := msg.String()
	switch s {
	case "y", "Y", "enter":
		cmd := m.confirmAction()
		m.confirmAction = nil
		m.confirmLabel = ""
		return m, cmd
	case "n", "N", "esc", "q":
		m.confirmAction = nil
		m.confirmLabel = ""
		m.status = "cancelled"
		return m, nil
	}
	return m, nil
}

func (m Model) withConfirm(label string, action func() tea.Cmd) Model {
	m.confirmAction = action
	m.confirmLabel = label
	m.status = ""
	return m
}

func (m Model) handleImageKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	s := msg.String()
	filt := m.filteredImages()
	switch {
	case contains(m.keys.Up, s):
		if m.curImg > 0 {
			m.curImg--
		}
		m.scroll = 0
		return m, m.loadCurrentTab()
	case contains(m.keys.Down, s):
		if m.curImg < len(filt)-1 {
			m.curImg++
		}
		m.scroll = 0
		return m, m.loadCurrentTab()
	case contains(m.keys.Remove, s):
		id := m.curImageID()
		return m.withConfirm(fmt.Sprintf("remove image %s?", id), func() tea.Cmd {
			return runAction(m.client, "image rm", id, func(c *client.WSLC, ctx context.Context, id string) error {
				return c.ImageRemove(ctx, id, true)
			})
		}), nil
	case contains(m.keys.Prune, s):
		return m.withConfirm("prune unused images?", func() tea.Cmd {
			return pruneAction(m.client, "image prune", func(c *client.WSLC, ctx context.Context) ([]byte, error) {
				return c.ImagePrune(ctx)
			})
		}), nil
	case contains(m.keys.Run, s):
		m.enterRunPrompt()
		return m, nil
	}
	return m, nil
}

func (m *Model) enterRunPrompt() {
	if m.curImg < 0 || m.curImg >= len(m.images) {
		m.status = "no image selected"
		return
	}
	img := m.images[m.curImg]
	m.inputMode = inputRun
	m.inputLabel = "run"
	m.inputBuf = runDefaultArgs(img)
	m.inputSeed = m.curImageID()
	m.status = ""
}

func (m *Model) enterBuildPrompt() {
	m.inputMode = inputBuild
	m.inputLabel = "build"
	m.inputBuf = ". -t "
	m.inputSeed = ""
	m.status = ""
}

func (m *Model) enterFilterPrompt() {
	m.inputMode = inputFilter
	m.inputLabel = "filter"
	m.inputBuf = ""
	m.inputSeed = ""
	m.status = ""
}

func runDefaultArgs(img client.Image) string {
	name := img.Repository
	if name == "" {
		name = "container"
	}
	return fmt.Sprintf("--name %s -d %s", name, imageRef(img))
}

func (m Model) handleInputKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	s := msg.String()
	switch s {
	case "esc", "ctrl+c":
		if m.inputMode == inputFilter {
			switch m.panel {
			case PanelContainers:
				m.filterCont = ""
			case PanelImages:
				m.filterImg = ""
			}
		}
		m.inputMode = inputNone
		m.inputBuf = ""
		m.inputLabel = ""
		m.inputSeed = ""
		m.status = "cancelled"
		return m, nil
	case "enter":
		return m.commitInput()
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
		return m, nil
	}
	if len(msg.Runes) > 0 {
		m.inputBuf += string(msg.Runes)
	}
	return m, nil
}

func (m Model) commitInput() (Model, tea.Cmd) {
	mode := m.inputMode
	buf := m.inputBuf
	seed := m.inputSeed
	m.inputMode = inputNone
	m.inputBuf = ""
	m.inputLabel = ""
	m.inputSeed = ""
	switch mode {
	case inputRun:
		args, err := parseRunArgs(buf)
		if err != nil {
			m.status = fmt.Sprintf("run: %v", err)
			return m, nil
		}
		image := seed
		if args.image != "" {
			image = args.image
		}
		if image == "" {
			m.status = "run: no image specified"
			return m, nil
		}
		return m, doRun(m.client, image, args)
	case inputBuild:
		path, tag := parseBuildArgs(buf)
		if path == "" {
			m.status = "build: usage: <path> -t <tag>"
			return m, nil
		}
		return m, doBuild(m.client, path, tag)
	case inputFilter:
		switch m.panel {
		case PanelContainers:
			m.filterCont = buf
		case PanelImages:
			m.filterImg = buf
		}
		m.curCont = 0
		m.curImg = 0
		return m, nil
	}
	return m, nil
}

type runArgs struct {
	name   string
	ports  []string
	detach bool
	rm     bool
	image  string
	cmd    []string
}

func parseRunArgs(s string) (runArgs, error) {
	fields := splitFields(s)
	var a runArgs
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		switch f {
		case "--rm":
			a.rm = true
		case "-d":
			a.detach = true
		case "--name":
			i++
			if i >= len(fields) {
				return a, fmt.Errorf("--name needs a value")
			}
			a.name = fields[i]
		case "-p":
			i++
			if i >= len(fields) {
				return a, fmt.Errorf("-p needs a value")
			}
			a.ports = append(a.ports, fields[i])
		case "--":
			a.cmd = append(a.cmd, fields[i+1:]...)
			return a, nil
		default:
			if a.image == "" {
				a.image = f
			} else {
				a.cmd = append(a.cmd, f)
			}
		}
	}
	return a, nil
}

func parseBuildArgs(s string) (path, tag string) {
	fields := splitFields(s)
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if f == "-t" || f == "--tag" {
			if i+1 < len(fields) {
				tag = fields[i+1]
				i++
			}
			continue
		}
		if path == "" {
			path = f
		}
	}
	return path, tag
}

func splitFields(s string) []string {
	var out []string
	var cur []rune
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case (r == ' ' || r == '\t') && !inQuote:
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = cur[:0]
			}
		default:
			cur = append(cur, r)
		}
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}

// loadCurrentTab dispatches the data load for whatever tab is active.
func (m Model) loadCurrentTab() tea.Cmd {
	switch m.panel {
	case PanelContainers, PanelMain:
		if m.panel == PanelContainers || m.panel == PanelMain {
			id := m.curContainerID()
			if id == "" {
				return nil
			}
			switch m.tab {
			case TabLogs:
				return loadLogs(m.client, id)
			case TabStats:
				return loadStats(m.client)
			case TabInspect:
				return loadInspect(m.client, id)
			}
		}
	case PanelImages:
		id := m.curImageID()
		if id != "" && m.tab == TabInspect {
			return loadImageInspect(m.client, id)
		}
	}
	return nil
}

func (m Model) filteredContainers() []client.Container {
	if m.filterCont == "" {
		return m.containers
	}
	var out []client.Container
	for _, c := range m.containers {
		if matchesFilter(c.Name, m.filterCont) || matchesFilter(c.Image, m.filterCont) {
			out = append(out, c)
		}
	}
	return out
}

func (m Model) filteredImages() []client.Image {
	if m.filterImg == "" {
		return m.images
	}
	var out []client.Image
	for _, im := range m.images {
		ref := imageRef(im)
		if matchesFilter(ref, m.filterImg) {
			out = append(out, im)
		}
	}
	return out
}

func matchesFilter(s, q string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(q))
}

func (m Model) curContainerID() string {
	filt := m.filteredContainers()
	if m.curCont < 0 || m.curCont >= len(filt) {
		return ""
	}
	c := filt[m.curCont]
	if c.Name != "" {
		return c.Name
	}
	return c.ID
}

func (m Model) curImageID() string {
	filt := m.filteredImages()
	if m.curImg < 0 || m.curImg >= len(filt) {
		return ""
	}
	i := filt[m.curImg]
	return imageRef(i)
}

func imageRef(i client.Image) string {
	if i.Repository != "" {
		return fmt.Sprintf("%s:%s", i.Repository, orDefault(i.Tag, "latest"))
	}
	return i.ID
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

// ---- View ---------------------------------------------------------------

func (m Model) View() string {
	if !m.ready {
		return "starting…\n"
	}

	// layout: 30% side | main
	sideW := m.width / 4
	if sideW < 24 {
		sideW = 24
	}
	mainW := m.width - sideW - 3

	side := m.renderSide(sideW, m.height-2)
	main := m.renderMain(mainW, m.height-2)
	status := m.renderStatus(m.width)

	body := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, side, "   ", main),
		status,
	)
	if m.inputMode != inputNone {
		prompt := m.renderInputPrompt()
		return lipgloss.JoinVertical(lipgloss.Left, body, prompt)
	}
	if m.confirmAction != nil {
		prompt := m.styles.Error.Render(fmt.Sprintf("%s [y/n]", m.confirmLabel))
		return lipgloss.JoinVertical(lipgloss.Left, body, prompt)
	}
	return body
}

func (m Model) renderInputPrompt() string {
	label := m.inputLabel
	hint := ""
	switch m.inputMode {
	case inputRun:
		hint = " [--name N] [-d] [--rm] [-p HOST:CTR] IMAGE [CMD...]"
	case inputBuild:
		hint = " <path> -t <tag>"
	case inputFilter:
		hint = " (esc to clear)"
	}
	return m.styles.Status.Render(
		fmt.Sprintf("%s%s> %s_", label, hint, m.inputBuf),
	)
}

func (m Model) renderSide(w, h int) string {
	containers := m.renderContainers(w/2-1, h)
	images := m.renderImages(w/2-1, h)
	return lipgloss.JoinHorizontal(lipgloss.Top, containers, " ", images)
}

func (m Model) renderContainers(w, h int) string {
	title := "containers"
	if m.filterCont != "" {
		title = fmt.Sprintf("containers /%s", m.filterCont)
	}
	if m.panel == PanelContainers {
		title = m.styles.TitleActive.Render(title)
	} else {
		title = m.styles.Title.Render(title)
	}
	body := m.renderContainerList(w, h)
	border := m.styles.Border
	if m.panel == PanelContainers {
		border = m.styles.BorderActive
	}
	return border.Width(w).Height(h).Render(title + "\n" + body)
}

func (m Model) renderContainerList(w, h int) string {
	filt := m.filteredContainers()
	if len(filt) == 0 {
		if m.filterCont != "" {
			return m.styles.Muted.Render(fmt.Sprintf("no match for %q", m.filterCont))
		}
		return m.styles.Muted.Render("none")
	}
	rows := make([]string, 0, len(filt))
	for i, c := range filt {
		marker := " "
		if i == m.curCont && m.panel == PanelContainers {
			marker = ">"
		}
		name := c.Name
		if name == "" {
			name = truncate(c.ID, 12)
		}
		state := c.StateString()
		line := fmt.Sprintf("%s %-12s %-10s", marker, truncate(name, 12), state)
		if i == m.curCont && m.panel == PanelContainers {
			line = m.styles.RowSel.Render(line)
		}
		rows = append(rows, line)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderImages(w, h int) string {
	title := "images"
	if m.filterImg != "" {
		title = fmt.Sprintf("images /%s", m.filterImg)
	}
	if m.panel == PanelImages {
		title = m.styles.TitleActive.Render(title)
	} else {
		title = m.styles.Title.Render(title)
	}
	body := m.renderImageList(w, h)
	border := m.styles.Border
	if m.panel == PanelImages {
		border = m.styles.BorderActive
	}
	return border.Width(w).Height(h).Render(title + "\n" + body)
}

func (m Model) renderImageList(w, h int) string {
	filt := m.filteredImages()
	if len(filt) == 0 {
		if m.filterImg != "" {
			return m.styles.Muted.Render(fmt.Sprintf("no match for %q", m.filterImg))
		}
		return m.styles.Muted.Render("none")
	}
	rows := make([]string, 0, len(filt))
	for i, im := range filt {
		marker := " "
		if i == m.curImg && m.panel == PanelImages {
			marker = ">"
		}
		name := im.Repository
		if name == "" {
			name = truncate(im.ID, 12)
		}
		tag := im.Tag
		if tag == "" {
			tag = "latest"
		}
		line := fmt.Sprintf("%s %-12s:%-10s", marker, truncate(name, 12), truncate(tag, 10))
		if i == m.curImg && m.panel == PanelImages {
			line = m.styles.RowSel.Render(line)
		}
		rows = append(rows, line)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderMain(w, h int) string {
	tabs := m.renderTabs(w)
	body := m.renderMainBody(w, h-2)
	border := m.styles.Border
	if m.panel == PanelMain {
		border = m.styles.BorderActive
	}
	return border.Width(w).Height(h).Render(tabs + "\n" + body)
}

func (m Model) renderTabs(w int) string {
	names := []string{"logs", "stats", "inspect", "config"}
	out := make([]string, 4)
	for i, n := range names {
		if i == int(m.tab) {
			out[i] = m.styles.TabActive.Render(n)
		} else {
			out[i] = m.styles.Tab.Render(n)
		}
	}
	return strings.Join(out, " ")
}

func (m Model) renderMainBody(w, h int) string {
	switch m.tab {
	case TabLogs:
		if m.logs == nil {
			return m.styles.Muted.Render("select a container")
		}
		if len(m.logs) == 0 {
			return m.styles.Muted.Render("no logs (last hour)")
		}
		// tail to fit height
		start := 0
		if len(m.logs) > h-1 {
			start = len(m.logs) - (h - 1)
		}
		return m.styles.Main.Render(scrollLines(strings.Join(m.logs[start:], "\n"), h, w, m.scroll))
	case TabStats:
		return m.renderStatsPanel(w, h)
	case TabInspect:
		if m.inspect == "" {
			return m.styles.Muted.Render("no inspect data")
		}
		return m.styles.Main.Render(scrollLines(m.inspect, h, w, m.scroll))
	case TabConfig:
		return m.styles.Muted.Render("config tab — press 'o' to open config (planned)")
	}
	return ""
}

func (m Model) renderStatus(w int) string {
	left := m.status
	if m.err != nil {
		left = m.styles.Error.Render(left)
	}
	help := m.styles.Help.Render(m.helpText())
	return lipgloss.JoinHorizontal(lipgloss.Left, left, "   ", help)
}

func (m Model) helpText() string {
	common := "q quit · tab panel · [/] tab · jk nav · R refresh · b build · / filter · pgup/pgdn scroll · mouse"
	switch m.panel {
	case PanelContainers:
		return common + " · s stop · S start · r restart · d rm · p prune"
	case PanelImages:
		return common + " · enter run · d rm · p prune"
	default:
		return common
	}
}

// ---- helpers ------------------------------------------------------------

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func truncateLines(s string, maxLines, maxWidth int) string {
	return scrollLines(s, maxLines, maxWidth, 0)
}

func scrollLines(s string, maxLines, maxWidth, scroll int) string {
	lines := strings.Split(s, "\n")
	total := len(lines)
	maxView := maxLines
	if maxView < 1 {
		maxView = 1
	}
	maxView-- // reserve a line for the position indicator
	start := total - maxView
	if scroll > 0 {
		start -= scroll
	}
	if start < 0 {
		start = 0
	}
	end := start + maxView
	if end > total {
		end = total
	}
	view := lines[start:end]
	for i, ln := range view {
		if maxWidth > 0 && len(ln) > maxWidth {
			view[i] = ln[:maxWidth-1] + "…"
		}
	}
	pos := fmt.Sprintf("[%d-%d/%d]", start+1, end, total)
	return strings.Join(view, "\n") + "\n" + pos
}

func (m *Model) scrollUp(n int) {
	m.scroll += n
}
func (m *Model) scrollDown(n int) {
	m.scroll -= n
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// ---- commands -----------------------------------------------------------

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func refresh(c *client.WSLC) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cs, errC := c.Containers(ctx, true)
		imgs, errI := c.Images(ctx)
		if errC != nil {
			return refreshMsg{err: errC}
		}
		if errI != nil {
			return refreshMsg{err: errI}
		}
		return refreshMsg{containers: cs, images: imgs}
	}
}

func loadLogs(c *client.WSLC, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		out, err := c.ContainerLogs(ctx, id, time.Hour)
		if err != nil {
			return logsMsg{err: err}
		}
		return logsMsg{lines: []string{string(out)}}
	}
}

func loadStats(c *client.WSLC) tea.Cmd {
	return sampleStats(c)
}

func sampleStats(c *client.WSLC) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stats, err := c.Stats(ctx)
		if err != nil {
			return statsMsg{err: err}
		}
		return statsMsg{stats: stats}
	}
}

func loadInspect(c *client.WSLC, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		out, err := c.ContainerInspect(ctx, id)
		if err != nil {
			return inspectMsg{err: err}
		}
		return inspectMsg{text: string(out)}
	}
}

func loadImageInspect(c *client.WSLC, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		out, err := c.ImageInspect(ctx, id)
		if err != nil {
			return inspectMsg{err: err}
		}
		return inspectMsg{text: string(out)}
	}
}

type actionFn func(c *client.WSLC, ctx context.Context, id string) error

func runAction(c *client.WSLC, action, id string, fn actionFn) tea.Cmd {
	if id == "" {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return actionDoneMsg{action: action, target: id, err: fn(c, ctx, id)}
	}
}

type pruneFn func(c *client.WSLC, ctx context.Context) ([]byte, error)

func pruneAction(c *client.WSLC, action string, fn pruneFn) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		out, err := fn(c, ctx)
		if err != nil {
			return actionDoneMsg{action: action, target: "", err: err}
		}
		return actionDoneMsg{action: action, target: strings.TrimSpace(string(out))}
	}
}

func doRun(c *client.WSLC, image string, a runArgs) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		out, err := c.Run(ctx, image, client.RunOptions{
			Name:   a.name,
			Ports:  a.ports,
			Detach: a.detach,
			Remove: a.rm,
			Cmd:    a.cmd,
		})
		return runResultMsg{out: out, err: err}
	}
}

func doBuild(c *client.WSLC, path, tag string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		out, err := c.Build(ctx, path, tag)
		if err != nil {
			return actionDoneMsg{action: "build", target: path, err: err}
		}
		return actionDoneMsg{action: "build", target: strings.TrimSpace(string(out))}
	}
}
