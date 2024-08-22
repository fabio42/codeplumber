package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
	"github.com/rs/zerolog/log"
)

// StatusLines represent the status lines
type StatusLines struct {
	tInput []textinput.Model
	sInput []textinput.Model

	notification  notification
	width, height int
	ui            *uiData
}

type notification struct {
	kind   string
	prompt string
	src    string
	ref    interface{}
}

func newInput(prompt, placeholder string, s lipgloss.Style) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 64
	ti.Width = 64
	ti.Prompt = prompt
	ti.Placeholder = fmt.Sprintf("%-64v", placeholder)
	ti.TextStyle = s
	ti.PlaceholderStyle = s
	return ti
}

// NewStatusLines returns a new StatusLines
func NewStatusLines(ui *uiData) *StatusLines {
	xSearch := make([]textinput.Model, len(supportedViews))
	xText := make([]textinput.Model, len(supportedViews))
	for i := range supportedViews {
		xSearch[i] = newInput("> ", "Search", ui.StatusLineSearchStyle())
		xText[i] = newInput(" ", "_", ui.StatusLineSearchStyle())
	}

	return &StatusLines{
		sInput: xSearch,
		tInput: xText,
		ui:     ui,
	}
}

// Init implement the tea.Model interface
func (m *StatusLines) Init() tea.Cmd {
	return nil
}

// Update implement the tea.Model interface
func (m *StatusLines) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug().Str("model", "tui").Str("func", "StatusLines.Update").Msgf("InputFocused: %v, InputType: %v, new msg: %v", m.ui.inputFocused, m.notification.kind, msg)
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch m.notification.kind {
		case "text", "reason":
			if key.Matches(msg, pagerKeys.Select) {
				m.tInput[m.ui.viewIdx].Blur()
				data := m.tInput[m.ui.viewIdx].Value()
				m.response(data, len(data) > 0)
			}

		case "confirm":
			switch {
			case key.Matches(msg, pagerKeys.Confirm):
				m.response("yes", true)
			case key.Matches(msg, pagerKeys.Decline):
				m.notification.kind = ""
				m.ui.inputFocused = false
			}

		case searchMsg:
			if key.Matches(msg, pagerKeys.Select) {
				m.sInput[m.ui.viewIdx].Blur()
				data := m.sInput[m.ui.viewIdx].Value()
				m.response(data, len(data) > 0)
			}

		default:
			if m.notification.kind == "error" {
				m.ui.inputFocused = false
				m.notification.kind = ""
			}
		}

	case notification:
		m.ui.inputFocused = true
		m.notification = msg
		switch msg.kind {
		case "reason":
			m.tInput[m.ui.viewIdx].Focus()
		case searchMsg:
			m.sInput[m.ui.viewIdx].Focus()
		}
	}

	m.sInput[m.ui.viewIdx], _ = m.sInput[m.ui.viewIdx].Update(msg)
	m.tInput[m.ui.viewIdx], _ = m.tInput[m.ui.viewIdx].Update(msg)
	return m, tea.Batch(cmds...)
}

// View implement the tea.Model interface
func (m *StatusLines) View() string {
	switch m.notification.kind {
	case "search":
		return m.sInput[m.ui.viewIdx].View()
	case "text", "reason":
		return lipgloss.NewStyle().Foreground(tint.Yellow()).Render(m.notification.prompt) + m.tInput[m.ui.viewIdx].View()
	case "confirm":
		return lipgloss.NewStyle().Foreground(tint.Yellow()).Render("CONFIRM: ") + m.notification.prompt + (" (y/n)")
	case errorMsg:
		return lipgloss.NewStyle().Foreground(tint.Red()).Render("ERROR: ") + m.notification.prompt + " (press any key to continue)"
	default:
		prompt := lipgloss.NewStyle().Bold(true).Foreground(tint.White()).Render("Path:")
		path := strings.Join(m.ui.path[0:m.ui.viewIdx+1], "/")
		return prompt + " /" + m.truncatePath(path, prompt+" /")
	}
}

func (m *StatusLines) response(data string, trigger bool) {
	m.ui.responseInput(
		m.notification.src,
		m.notification.kind,
		data,
		m.notification.ref,
		trigger,
	)
	// Clear notification
	m.ui.inputFocused = false
	m.notification = notification{}
}

func (m *StatusLines) truncatePath(path, padding string) string {
	rpath := []rune(path)
	lPadding := lipgloss.Width(padding)
	pathLen := len(rpath)

	for m.width-lipgloss.Width(string(rpath))-lPadding-2 < 0 {
		if idx := len(rpath) - 1; idx > 0 {
			rpath = rpath[len(rpath)-idx:]
		} else {
			break
		}
	}

	if pathLen > len(rpath) {
		rpath[0] = '…'
	}
	return string(rpath)
}
