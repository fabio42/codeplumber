package tui

import (
	"codeplumber/models/table"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

const (
	pipelinesFilter = "pipelinesFilter"
)

// PipelinesTable is the model for the pipelines table
type PipelinesTable struct {
	*table.Model
	help          help.Model
	name          string
	width, height int
	ui            *uiData
	allRows       []table.Row
}

// NewPipelinesTable returns a new PipelinesTable
func NewPipelinesTable(ui *uiData) *PipelinesTable {
	t := table.New()
	t.SetStyles(ui.getTablePatchedStyle())
	return &PipelinesTable{
		Model: &t,
		name:  "codepipelines",
		ui:    ui,
		help:  help.New(),
	}
}

// SetColumns set the columns of the table
func (m *PipelinesTable) SetColumns(width int) {
	width = width - 5
	triggerSize := percent(width, 25, 20)
	statusSize := percent(width, 25, 10)
	lastExecutionSize := percent(width, 40, 22)
	nameSize := width - statusSize - lastExecutionSize - triggerSize

	cols := make([]table.Column, 4)
	cols[0] = table.Column{Title: "CodePipeline Name", Width: nameSize}
	cols[1] = table.Column{Title: "User", Width: triggerSize}
	cols[2] = table.Column{Title: "Status", Width: statusSize}
	cols[3] = table.Column{Title: "Last execution", Width: lastExecutionSize}
	m.Model.SetColumns(cols)
	m.Focus()
}

// SetWidth set the width of the table
func (m *PipelinesTable) SetWidth(width int) {
	m.SetColumns(width)
	m.Model.SetWidth(width)
}

// Init implement the tea.Model interface
func (m *PipelinesTable) Init() tea.Cmd {
	return nil
}

// Update implement the tea.Model interface
func (m *PipelinesTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
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
		case key.Matches(msg, allKeys.Select):
			if len(m.SelectedRow()) > 0 {
				m.ui.changeView(pipelinesView, pipelineView, m.SelectedRow()[0])
			}

		case key.Matches(msg, codePipelineKeys.Start):
			if len(m.SelectedRow()) > 0 {
				m.ui.confirm(pipelineStart, "Start this CodePipeline?", nil)
			}

		case key.Matches(msg, allKeys.Refresh):
			m.refresh()

		case key.Matches(msg, allKeys.Search):
			m.ui.search(pipelinesFilter)

		case key.Matches(msg, allKeys.Browse):
			m.browse()
		}

	case redraw:
		m.UpdateViewport()

	case tuiMsg:
		log.Debug().Str("model", "tui").Str("func", "PipelinesTable.Update").Msgf("tuiMsg class: %v, id: %v, trigger: %v, data: %v", msg.class, msg.id, msg.trigger, msg.data)
		switch msg.class {
		case response:
			switch msg.src {
			case pipelineStart:
				if msg.trigger {
					m.start(m.SelectedRow()[0])
					m.refresh()
				}
			case pipelinesFilter:
				go m.filterOperations(msg.data.(string))
			}
		case viewUpdate:
			rows := msg.data.([]table.Row)
			m.SetColumns(m.width)
			m.SetRows(rows)
			if len(m.allRows) == 0 {
				m.allRows = rows
			}

		default:
			m.SetColumns(m.width)
			m.SetRows(msg.data.([]table.Row))
		}
	}

	*m.Model, _ = m.Model.Update(msg)
	return m, tea.Batch(cmds...)
}

// View implement the tea.Model interface
func (m *PipelinesTable) View() string {
	var help string
	if m.ui.help {
		help = m.helpViewFull()
	} else {
		help = m.helpView()
	}

	return fmt.Sprintf("%s\n%s", m.Model.View(), help)
}

func (m *PipelinesTable) helpView() string {
	return m.help.ShortHelpView([]key.Binding{
		allKeys.Select,
		allKeys.Previous,
		allKeys.Search,
		codePipelineKeys.Start,
		allKeys.Help,
	})
}

func (m *PipelinesTable) helpViewFull() string {
	return m.help.FullHelpView([][]key.Binding{
		{
			allKeys.Up,
			allKeys.Down,
			allKeys.Select,
			allKeys.Previous,
		},
		{
			allKeys.Browse,
			allKeys.Search,
			codePipelineKeys.Start,
		},
		{
			allKeys.Refresh,
			allKeys.Quit,
			allKeys.Help,
		},
	})
}
