package tui

import (
	"strings"

	"github.com/fabio42/codeplumber/models/table"

	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

var containerStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

func (c *uiData) getTablePatchedStyle() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true).
		BorderBottom(true).
		Bold(true)

	s.RenderCell = func(_ table.Model, value string, position table.CellPosition) string {
		if position.IsRowSelected {
			return s.Cell.
				Foreground(lipgloss.Color("229")).
				Background(tint.Purple()).
				Render(value)
		}

		color := c.statusColor(value)
		if color != nil {
			return s.Cell.Foreground(color).Render(value)
		}
		return s.Cell.Render(value)
	}

	return s
}

// statusColor will return a colored lipgloss style string based on the Status
func (c *uiData) statusColor(status string) lipgloss.TerminalColor {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "inprogress", "waiting":
		return tint.Blue()
	case "enabled", "succeeded":
		return tint.Green()
	case "disabled", "failed":
		return tint.Red()
	case "stopped", "unknown":
		return tint.Yellow()
	case "resuming":
		return tint.BrightPurple()
	default:
		return nil
	}
}

func (c *uiData) StatusLineSearchStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(tint.Yellow())
}
