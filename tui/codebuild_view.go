package tui

import (
	"fmt"
	"strings"

	"github.com/fabio42/codeplumber/models/table"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

// CodeBuildTable represent a AWS CodeBuild details
type CodeBuildTable struct {
	*table.Model
	name          string
	buildID       string
	pipelineName  string
	stageName     string
	width, height int
	ui            *uiData
	help          help.Model
}

// NewCodeBuildTable returns a new CodeBuildTable
func NewCodeBuildTable(ui *uiData) *CodeBuildTable {
	t := table.New()
	t.SetStyles(ui.getTablePatchedStyle())
	return &CodeBuildTable{
		Model: &t,
		ui:    ui,
		help:  help.New(),
	}
}

// SetColumns set the columns of the CodeBuildTable
func (m *CodeBuildTable) SetColumns(width int) {
	cols := make([]table.Column, 4)

	width = width - 5
	componentSize := percent(width, 40, 25)
	valueSize := width - componentSize

	cols[0] = table.Column{Title: "CodeBuild option", Width: componentSize}
	cols[1] = table.Column{Title: "Value", Width: valueSize}
	m.Model.SetColumns(cols)
	m.Focus()
}

// SetWidth set the width of the CodeBuildTable
func (m *CodeBuildTable) SetWidth(width int) {
	m.SetColumns(width)
	m.Model.SetWidth(width)
}

// Init implement the tea.Model interface
func (m *CodeBuildTable) Init() tea.Cmd {
	return nil
}

// Update implement the tea.Model interface
func (m *CodeBuildTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.ui.updatPath(m.name)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMarginHeight := 4
		if m.ui.help {
			verticalMarginHeight += helpFullHeiggt
		} else {
			verticalMarginHeight += helpHeight
		}

		m.width = msg.Width - 2
		m.height = msg.Height - verticalMarginHeight
		m.SetHeight(m.height)
		m.SetWidth(m.width)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, allKeys.Previous):
			m.ui.previousViewWithRefresh()

		case key.Matches(msg, allKeys.Refresh):
			m.refresh(m.buildID)

		case key.Matches(msg, codebuildKeys.Log):
			m.ui.changeView(codebuildView, logView, PagerSelector{name: m.name})

		case key.Matches(msg, allKeys.Browse):
			m.browse()

		case key.Matches(msg, allKeys.Select):
			s := m.SelectedRow()
			if strings.Contains(s[0], "URL") {
				m.ui.changeView(codebuildView, logView, PagerSelector{name: m.name})
			} else {
				m.ui.changeView(codebuildView, buildspecView, PagerSelector{name: m.name, content: s[1]})
			}
		}

	case tuiMsg:
		log.Debug().Str("model", "tui").Str("func", "CodeBuildTable.Update").Msgf("tuiMsg class: %v, id: %v, trigger: %v, data: %v", msg.class, msg.id, msg.trigger, msg.data)
		switch msg.class {
		case viewChange:
			m.SetRows([]table.Row{})
			payload := msg.data.(PipelineResource)
			m.pipelineName = payload.PipelineName
			m.stageName = payload.StageName
			m.buildID = payload.ExternalExecutionID
			m.SetColumns(m.width)
			m.refresh(m.buildID)

		case viewUpdate:
			rows := msg.data.([]table.Row)
			m.SetColumns(m.width)
			m.SetRows(rows)
		}
	}

	*m.Model, _ = m.Model.Update(msg)
	return m, nil
}

// View implement the tea.Model interface
func (m *CodeBuildTable) View() string {
	var help string
	if m.ui.help {
		help = m.helpViewFull()
	} else {
		help = m.helpView()
	}

	return fmt.Sprintf("%s\n%s", m.Model.View(), help)
}

func (m *CodeBuildTable) helpView() string {
	return m.help.ShortHelpView([]key.Binding{
		allKeys.Select,
		allKeys.Previous,
		codebuildKeys.Log,
	})
}

func (m *CodeBuildTable) helpViewFull() string {
	return m.help.FullHelpView([][]key.Binding{
		{
			allKeys.Up,
			allKeys.Down,
			allKeys.Select,
			allKeys.Previous,
		},
		{
			codebuildKeys.Log,
			allKeys.Browse,
		},
		{
			allKeys.Refresh,
			allKeys.Quit,
			allKeys.Help,
		},
	})
}
