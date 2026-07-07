package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gregorypilar/lazywslcontainer/internal/client"
	"github.com/gregorypilar/lazywslcontainer/internal/tui"
)

func main() {
	c := client.New("wslc")
	if _, err := c.Ping(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "lazywslcontainer:", err)
		os.Exit(1)
	}
	p := tea.NewProgram(tui.New(c), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazywslcontainer:", err)
		os.Exit(1)
	}
}
