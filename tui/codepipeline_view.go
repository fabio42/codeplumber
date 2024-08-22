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

const (
	pipelineStart     = "codepipelineStart"
	stageRestart      = "codepipelineStageRestart"
	transitionEnable  = "codepipelineTransitionEnable"
	transitionDisable = "codepipelineTransitionDisable"
)

// PipelineTable represent a AWS CodePipeline details
type PipelineTable struct {
	*table.Model
	name          string
	width, height int
	ui            *uiData
	help          help.Model
	confirm       struct {
		wait    bool
		element interface{}
	}
}

// NewPipelineTable returns a new PipelineTable
func NewPipelineTable(ui *uiData) *PipelineTable {
	t := table.New()
	t.SetStyles(ui.getTablePatchedStyle())
	return &PipelineTable{
		Model: &t,
		ui:    ui,
		help:  help.New(),
	}
}

// SetColumns set the columns of the table
func (m *PipelineTable) SetColumns(width int) {
	cols := make([]table.Column, 5)

	width = width - 5
	typeSize := percent(width, 20, 40)
	// Not usefull for user, hidding it
	stageSize := 0
	statusSize := percent(width, 20, 12)
	lastExecutionSize := percent(width, 20, 26)
	nameSize := width - typeSize - stageSize - statusSize - lastExecutionSize

	cols[0] = table.Column{Title: "Ressource Name", Width: nameSize}
	cols[1] = table.Column{Title: "Stage Type", Width: typeSize}
	cols[2] = table.Column{Title: "Stage Name", Width: stageSize}
	cols[3] = table.Column{Title: "Status", Width: statusSize}
	cols[4] = table.Column{Title: "Last execution", Width: lastExecutionSize}
	m.Model.SetColumns(cols)
	m.Focus()
}

// SetWidth set the width of the table
func (m *PipelineTable) SetWidth(width int) {
	m.SetColumns(width)
	m.Model.SetWidth(width)
}

// Init implement the tea.Model interface
func (m *PipelineTable) Init() tea.Cmd {
	return nil
}

// Update implement the tea.Model interface
func (m *PipelineTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, allKeys.Previous):
			m.ui.previousViewWithRefresh()

		case key.Matches(msg, allKeys.Refresh):
			m.refresh()

		case key.Matches(msg, codePipelineKeys.Start):
			m.ui.confirm(pipelineStart, "Start this CodePipeline?", nil)

		case key.Matches(msg, codePipelineKeys.ReStart):
			s := m.SelectedRow()
			if strings.HasPrefix(s[0], separatorStage) {
				_, p, _ := m.selectComponent(s)
				if p.Status == "Failed" {
					m.ui.confirm(stageRestart, "Restart CodePipeline Stage?", p)
				} else {
					m.ui.errorMsg(pipelineView, "Can't restart a successful CodePipeline stage.")
				}
			}

		case key.Matches(msg, codePipelineKeys.ToggleTransition):
			transition := m.SelectedRow()
			if strings.HasPrefix(transition[0], separatorTransition) {
				if transition[3] == "Enabled" {
					m.ui.requestInput(transitionDisable, "reason", "DISABLE TRANSITION: short description (empty to cancel):", transition)
				} else {
					m.ui.confirm(transitionEnable, "Enable transition?", transition)
				}
			}

		case key.Matches(msg, allKeys.Select):
			s := m.SelectedRow()
			if strings.HasPrefix(s[0], separatorStage) {
				stageType, d, err := m.selectComponent(s)
				log.Debug().Str("model", "tui").Str("func", "PipelineTable.Update").Msgf("selected stageType: %v, data: %v", stageType, d)
				if err != nil {
					go m.ui.errorMsg(pipelineView, "Execution not ready... refreshing.")
					m.refresh()
				} else {
					m.ui.changeView(pipelineView, stageType, d)
				}
			}

		case key.Matches(msg, allKeys.Browse):
			m.browse()
		}

	case tuiMsg:
		log.Debug().Str("model", "tui").Str("func", "PipelineTable.Update").Msgf("tuiMsg class: %v, id: %v, trigger: %v, data: %v", msg.class, msg.id, msg.trigger, msg.data)
		var err error
		switch msg.class {
		case viewChange:
			m.name = msg.data.(string)
			m.SetColumns(m.width)
			m.SetRows([]table.Row{})
			m.refresh()

		case viewUpdate:
			rows := msg.data.([]table.Row)
			m.SetColumns(m.width)
			m.SetRows(rows)

		case response:
			if msg.trigger {
				switch msg.src {
				case transitionDisable:
					err = m.toggleTransition("disable", msg.data.(string))
				case transitionEnable:
					// Opiniated choice to start the pipeline before enabling the transition
					// This will ensure that the stage after the transition will have latest version of the previous stage input
					// Enabling a transition will automatically trigger the next stage if an output is waiting in queue,
					// And we don't want to trigger the next stage with an outdated input.
					// Open to better solution or introducing another key switch if this is cause issues.
					err = m.toggleTransition("enable", "")
				case stageRestart:
					if msg.trigger {
						m.restartStage(msg.reference.(PipelineResource))
					}
				case pipelineStart:
					if msg.trigger {
						m.start()
					}
				}
			}
			if err != nil {
				m.ui.errorMsg(pipelineView, err.Error())
			}
			m.refresh()
		}

	case refresh:
		m.refresh()
	}

	*m.Model, _ = m.Model.Update(msg)
	return m, tea.Batch(cmds...)
}

// View implement the tea.Model interface
func (m *PipelineTable) View() string {
	var help string
	if m.ui.help {
		help = m.helpViewFull()
	} else {
		help = m.helpView()
	}

	return fmt.Sprintf("%s\n%s", m.Model.View(), help)
}

func (m *PipelineTable) selectComponent(row table.Row) (string, PipelineResource, error) {
	var d PipelineResource
	var err error

	actionName := strings.Split(row[0], " ")[1]
	stageName := row[2]
	k := row[1]
	stageType := strings.ToLower(strings.Split(k, "/")[0])
	stageSection := strings.Split(k, "/")[1]

	switch stageType {
	case codebuildView:
		d, err = m.getCodebuildResource(m.name, stageName, stageSection, actionName)
		if err != nil {
			return "", d, err
		}
	default:
		d = PipelineResource{
			PipelineName: m.name,
			StageName:    actionName,
			ActionName:   stageSection,
		}
	}
	return stageType, d, nil
}

func (m *PipelineTable) helpView() string {
	return m.help.ShortHelpView([]key.Binding{
		allKeys.Select,
		allKeys.Previous,
		codePipelineKeys.Start,
		codePipelineKeys.ReStart,
		codePipelineKeys.ToggleTransition,
		allKeys.Help,
	})
}

func (m *PipelineTable) helpViewFull() string {
	return m.help.FullHelpView([][]key.Binding{
		{
			allKeys.Up,
			allKeys.Down,
			allKeys.Select,
			allKeys.Previous,
		},
		{
			allKeys.Browse,
			codePipelineKeys.Start,
			codePipelineKeys.ReStart,
			codePipelineKeys.ToggleTransition,
		},
		{
			allKeys.Refresh,
			allKeys.Quit,
			allKeys.Help,
		},
	})
}
