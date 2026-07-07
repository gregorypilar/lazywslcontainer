package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gregorypilar/lazywslcontainer/internal/client"
)

const statsMaxSamples = 60

type containerStats struct {
	name string
	cpu  []float64
	mem  []float64
}

func (s *containerStats) push(cpu, mem float64) {
	s.cpu = append(s.cpu, cpu)
	s.mem = append(s.mem, mem)
	if len(s.cpu) > statsMaxSamples {
		s.cpu = s.cpu[len(s.cpu)-statsMaxSamples:]
	}
	if len(s.mem) > statsMaxSamples {
		s.mem = s.mem[len(s.mem)-statsMaxSamples:]
	}
}

func parseCPUPerc(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseMemBytes(s string) float64 {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(parts[0], 64)
	if len(parts) > 1 {
		switch strings.ToUpper(parts[1]) {
		case "B":
			v *= 1
		case "KIB", "KB":
			v *= 1024
		case "MIB", "MB":
			v *= 1024 * 1024
		case "GIB", "GB":
			v *= 1024 * 1024 * 1024
		case "TIB", "TB":
			v *= 1024 * 1024 * 1024 * 1024
		}
	}
	return v
}

// sparklineMulti renders a sparkline with `height` rows using 8-step blocks.
// The vertical resolution is height*8. Each row shows the lower part of the
// range first (bottom row = lowest values).
func sparklineMulti(vals []float64, width, height int, min, max float64, color lipgloss.Color) string {
	if len(vals) == 0 {
		return strings.Repeat("\n", height-1) + strings.Repeat(" ", width)
	}
	if width > len(vals) {
		width = len(vals)
	}
	start := len(vals) - width
	window := vals[start:]
	rng := max - min
	if rng == 0 {
		rng = 1
	}
	steps := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	rows := make([]string, height)
	for row := 0; row < height; row++ {
		topFrac := float64(height-row) / float64(height)
		botFrac := float64(height-row-1) / float64(height)
		out := make([]rune, 0, width)
		for _, v := range window {
			norm := (v - min) / rng
			if norm <= botFrac {
				out = append(out, ' ')
				continue
			}
			if norm >= topFrac {
				out = append(out, '█')
				continue
			}
			localFrac := (norm - botFrac) / (topFrac - botFrac)
			idx := int(localFrac * float64(len(steps)-1))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(steps) {
				idx = len(steps) - 1
			}
			out = append(out, steps[idx])
		}
		rows[row] = string(out)
	}
	styled := make([]string, height)
	for i, r := range rows {
		styled[i] = lipgloss.NewStyle().Foreground(color).Render(r)
	}
	return strings.Join(styled, "\n")
}

func (m *Model) ingestStats(stats []client.Stat) {
	if m.statsHistory == nil {
		m.statsHistory = make(map[string]*containerStats)
	}
	seen := make(map[string]bool, len(stats))
	for _, s := range stats {
		seen[s.ID] = true
		cs, ok := m.statsHistory[s.ID]
		if !ok {
			cs = &containerStats{name: s.Name}
			m.statsHistory[s.ID] = cs
		}
		if s.Name != "" {
			cs.name = s.Name
		}
		cs.push(parseCPUPerc(s.CPUPerc), parseMemBytes(s.MemUsage))
	}
	for id := range m.statsHistory {
		if !seen[id] {
			delete(m.statsHistory, id)
		}
	}
}

var (
	cpuColor   = lipgloss.Color("51")  // cyan
	memColor   = lipgloss.Color("213") // magenta
	nameColor  = lipgloss.Color("252")
	labelColor = lipgloss.Color("241")
)

func (m Model) renderStatsPanel(w, h int) string {
	if len(m.statsHistory) == 0 {
		return m.styles.Muted.Render("no stats yet — waiting for samples")
	}

	sparkW := w - 30
	if sparkW < 20 {
		sparkW = 20
	}
	sparkH := 3

	blocks := make([]string, 0, len(m.statsHistory))
	idx := 0
	for _, cs := range m.statsHistory {
		idx++
		curCPU := lastVal(cs.cpu)
		curMem := lastVal(cs.mem)
		minCPU, maxCPU := minMax(cs.cpu)
		minMem, maxMem := minMax(cs.mem)

		header := fmt.Sprintf(" %s  cpu %s%%  mem %s",
			lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(truncate(cs.name, 20)),
			lipgloss.NewStyle().Foreground(cpuColor).Render(fmt.Sprintf("%.2f", curCPU)),
			lipgloss.NewStyle().Foreground(memColor).Render(fmtBytes(curMem)),
		)

		cpuLabel := lipgloss.NewStyle().Foreground(labelColor).Render("cpu")
		cpuSpark := sparklineMulti(cs.cpu, sparkW, sparkH, 0, 100, cpuColor)
		memLabel := lipgloss.NewStyle().Foreground(labelColor).Render("mem")
		memSpark := sparklineMulti(cs.mem, sparkW, sparkH, minMem, maxMem, memColor)

		block := strings.Join([]string{
			header,
			fmt.Sprintf("%s %s", cpuLabel, indentMultiline(cpuSpark, 4)),
			fmt.Sprintf("%s %s", memLabel, indentMultiline(memSpark, 4)),
			fmt.Sprintf("  %s  %s",
				lipgloss.NewStyle().Foreground(labelColor).Render(fmt.Sprintf("cpu %5.1f%%–%5.1f%%", minCPU, maxCPU)),
				lipgloss.NewStyle().Foreground(labelColor).Render(fmt.Sprintf("mem %s–%s", fmtBytes(minMem), fmtBytes(maxMem))),
			),
		}, "\n")
		blocks = append(blocks, block)
		if idx < len(m.statsHistory) {
			blocks = append(blocks, lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Render(strings.Repeat("─", w-2)))
		}
	}

	out := strings.Join(blocks, "\n")
	lines := strings.Split(out, "\n")
	if len(lines) > h {
		lines = lines[len(lines)-h:]
	}
	return m.styles.Main.Render(strings.Join(lines, "\n"))
}

func indentMultiline(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i := range lines {
		if i == 0 {
			continue
		}
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

func minMax(v []float64) (float64, float64) {
	if len(v) == 0 {
		return 0, 0
	}
	min, max := v[0], v[0]
	for _, x := range v {
		if x < min {
			min = x
		}
		if x > max {
			max = x
		}
	}
	return min, max
}

func lastVal(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return v[len(v)-1]
}

func fmtBytes(b float64) string {
	const (
		KB = 1024.0
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.2fTiB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGiB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMiB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKiB", b/KB)
	default:
		return fmt.Sprintf("%.0fB", b)
	}
}
