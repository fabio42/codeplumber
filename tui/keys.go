package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up               key.Binding
	Down             key.Binding
	Select           key.Binding
	Previous         key.Binding
	Help             key.Binding
	Quit             key.Binding
	Refresh          key.Binding
	Filter           key.Binding
	Search           key.Binding
	Log              key.Binding
	Browse           key.Binding
	NextTint         key.Binding
	PrevTint         key.Binding
	Start            key.Binding
	ToggleTransition key.Binding
	ReStart          key.Binding
	Confirm          key.Binding
	Decline          key.Binding
}

var allKeys = keyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Select:   key.NewBinding(key.WithKeys("right", "l", "enter"), key.WithHelp("→/l/retrun", "select")),
	Previous: key.NewBinding(key.WithKeys("left", "h", "backspace"), key.WithHelp("←/h/backspace", "previous")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Filter:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Browse:   key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "open in browser")),
	NextTint: key.NewBinding(key.WithKeys("]"), key.WithHelp("t", "next tint")),
	PrevTint: key.NewBinding(key.WithKeys("["), key.WithHelp("T", "previous tint")),
	Start:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
}

var codebuildKeys = keyMap{
	Log:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "log")),
	Start: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "restart failed CodeBuild")),
}

var codePipelineKeys = keyMap{
	Start:            key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start CodePipeline")),
	ReStart:          key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "restart failed stage")),
	ToggleTransition: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "toggle transition")),
}

var pagerKeys = keyMap{
	Select:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
	Decline: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
}

var helpLeft = []key.Binding{
	allKeys.Up,
	allKeys.Down,
	allKeys.Select,
	allKeys.Previous,
}
