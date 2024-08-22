package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type defaultTint struct{}

func (t defaultTint) DisplayName() string {
	return "default"
}

func (t defaultTint) ID() string {
	return "default"
}

func (t defaultTint) About() string {
	return `Tint: default`
}

// Fg returns the recommended default foreground color for this tint.
func (t *defaultTint) Fg() lipgloss.TerminalColor {
	return lipgloss.NoColor{}
}

// Bg returns the recommended default background color for this tint.
func (t *defaultTint) Bg() lipgloss.TerminalColor {
	return lipgloss.Color("8")
	// return lipgloss.NoColor{}
}

// SelectionBg returns the recommended background color for selected text.
func (t *defaultTint) SelectionBg() lipgloss.TerminalColor {
	return lipgloss.NoColor{}
}

// Cursor returns the recommended color for the cursor.
func (t *defaultTint) Cursor() lipgloss.TerminalColor {
	return lipgloss.NoColor{}
}

func (t *defaultTint) BrightBlack() lipgloss.TerminalColor {
	return lipgloss.Color("8")
}

func (t *defaultTint) BrightBlue() lipgloss.TerminalColor {
	return lipgloss.Color("12")
}

func (t *defaultTint) BrightCyan() lipgloss.TerminalColor {
	return lipgloss.Color("14")
}

func (t *defaultTint) BrightGreen() lipgloss.TerminalColor {
	return lipgloss.Color("10")
}

func (t *defaultTint) BrightPurple() lipgloss.TerminalColor {
	return lipgloss.Color("13")
}

func (t *defaultTint) BrightRed() lipgloss.TerminalColor {
	return lipgloss.Color("9")
}

func (t *defaultTint) BrightWhite() lipgloss.TerminalColor {
	return lipgloss.Color("15")
}

func (t *defaultTint) BrightYellow() lipgloss.TerminalColor {
	return lipgloss.Color("11")
}

func (t *defaultTint) Black() lipgloss.TerminalColor {
	return lipgloss.Color("0")
}

func (t *defaultTint) Blue() lipgloss.TerminalColor {
	return lipgloss.Color("4")
}

func (t *defaultTint) Cyan() lipgloss.TerminalColor {
	return lipgloss.Color("6")
}

func (t *defaultTint) Green() lipgloss.TerminalColor {
	return lipgloss.Color("2")
}

func (t *defaultTint) Purple() lipgloss.TerminalColor {
	return lipgloss.Color("5")
}

func (t *defaultTint) Red() lipgloss.TerminalColor {
	return lipgloss.Color("1")
}

func (t *defaultTint) White() lipgloss.TerminalColor {
	return lipgloss.Color("7")
}

func (t *defaultTint) Yellow() lipgloss.TerminalColor {
	return lipgloss.Color("3")
}
