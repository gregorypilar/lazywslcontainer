package tui

import "github.com/charmbracelet/lipgloss"

// Palette mirrors lazydocker's dark, low-saturation look.
type Styles struct {
	Border       lipgloss.Style
	BorderActive lipgloss.Style

	Title       lipgloss.Style
	TitleActive lipgloss.Style

	Row      lipgloss.Style
	RowSel   lipgloss.Style
	RowAlt   lipgloss.Style
	RowState lipgloss.Style

	Tab       lipgloss.Style
	TabActive lipgloss.Style

	Status  lipgloss.Style
	Help    lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
	Muted   lipgloss.Style

	Main lipgloss.Style
}

func DefaultStyles() Styles {
	border := lipgloss.RoundedBorder()

	b := func(active bool) lipgloss.Style {
		s := lipgloss.NewStyle().Border(border, true)
		if active {
			s = s.BorderForeground(lipgloss.Color("205"))
		} else {
			s = s.BorderForeground(lipgloss.Color("240"))
		}
		return s.Padding(0, 1)
	}

	return Styles{
		Border:       b(false),
		BorderActive: b(true),

		Title:       lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true),
		TitleActive: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),

		Row:      lipgloss.NewStyle(),
		RowSel:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		RowAlt:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		RowState: lipgloss.NewStyle().Foreground(lipgloss.Color("36")),

		Tab:       lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 2),
		TabActive: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Padding(0, 2).Underline(true),

		Status:  lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		Help:    lipgloss.NewStyle().Foreground(lipgloss.Color("239")),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")),

		Main: lipgloss.NewStyle().Padding(0, 1),
	}
}
