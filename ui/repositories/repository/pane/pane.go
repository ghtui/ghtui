package pane

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"ghtui/ghtui/ui/common"
)

type Model struct {
	Viewport viewport.Model
	Active   bool
	Width    int
	Height   int
}

func NewModel(width int, height int, active bool) Model {
	return Model{
		Viewport: viewport.Model{
			Width:  width - 2,
			Height: height - 2,
		},
		Active: active,
		Width:  width,
		Height: height,
	}
}

func (m Model) View() string {
	var style lipgloss.Style
	if m.Active {
		style = common.PaneActiveStyle(m.Width, m.Height)
	} else {
		style = common.PaneInactiveStyle(m.Width, m.Height)
	}
	return style.Render(m.Viewport.View())
}
