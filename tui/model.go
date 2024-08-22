package tui

import (
	awsqueries "codeplumber/aws"
	"codeplumber/cmd/vcr"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
	"github.com/rs/zerolog/log"
)

const (
	helpHeight     = 1
	helpFullHeiggt = 4
	// Message class
	input      = "input"
	response   = "inputResponse"
	previous   = "previous"
	errorMsg   = "error"
	searchMsg  = "search"
	viewUpdate = "viewData"
	viewChange = "viewChange"
	rowsMsg    = "rows"
	// View class
	pipelinesView = "pipelines"
	pipelineView  = "pipeline"
	codebuildView = "codebuild"
	buildspecView = "buildspec"
	logView       = "log"
)

var (
	config         Config
	supportedViews = []string{pipelinesView, pipelineView, codebuildView, buildspecView, logView}
	supportFilter  = []string{pipelinesView}
)

// Config is the configuration for the TUI
type Config struct {
	AwsConfig       aws.Config
	Recorder        *vcr.Recorder
	RecordDir       string
	NameFilter      string
	NameFilterExtra string
	TagFilter       map[string]string
	Theme           string
	Mode            struct {
		Record, Replay bool
	}

	awsAccountID string
}

type tuiMsg struct {
	src       string
	id        string
	class     string
	trigger   bool
	reference interface{}
	data      interface{}
}

type redraw struct{}

type refresh struct{}

// Model is the main model for the TUI
type Model struct {
	statusLine      *StatusLines
	pipelinesTable  *PipelinesTable
	pipelineDetail  *PipelineTable
	codeBuildDetail *CodeBuildTable
	pager           *Pager
	spinner         spinner.Model

	quitting bool
	ui       *uiData

	isError       string
	width, height int
}

// NewModel creates a new model
func NewModel(cfg Config) *Model {
	// config is expected to be a global variable
	config = cfg
	var err error

	// Recorder is development tool to record and replay data
	config.Recorder = vcr.NewRecorder(config.Mode.Record, config.Mode.Replay, config.RecordDir)

	if config.Mode.Replay {
		s, _ := config.Recorder.Play("accountID", config.awsAccountID)
		config.awsAccountID = s.(string)
	} else {
		config.awsAccountID = awsqueries.GetAwsAccountID(config.AwsConfig)
		err = config.Recorder.Record("accountID", config.awsAccountID)
	}
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to record accountID: %v", err)
	}

	// Set the main model message channel
	selectChan := make(chan tuiMsg) // buffered or unbuffered?

	// Set tui theme
	tint.NewDefaultRegistry()
	tint.Register(&defaultTint{})
	if config.Theme != "" {
		tint.SetTintID(config.Theme)
	} else {
		tint.SetTintID("default")
	}

	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	s.Style = lipgloss.NewStyle().Foreground(tint.BrightBlue())

	ui := &uiData{
		selection: selectChan,
		dataCache: DataCache{
			pipelines:  make(map[string]awsqueries.Pipeline),
			codebuilds: make(map[string]awsqueries.CodebuildData),
		},
		views: []string{"pipelines"},
		path:  make([]string, len(supportedViews)),
	}

	return &Model{
		statusLine:      NewStatusLines(ui),
		pipelinesTable:  NewPipelinesTable(ui),
		pipelineDetail:  NewPipelineTable(ui),
		codeBuildDetail: NewCodeBuildTable(ui),
		pager:           NewPager(ui),
		spinner:         s,
		ui:              ui,
	}
}

// Init initializes the parent model
func (m *Model) Init() tea.Cmd {
	log.Debug().Str("model", "tui").Str("func", "Model.Init").Msg("new model")
	return tea.Batch(
		tea.EnterAltScreen,
		tea.ClearScreen,
		m.spinner.Tick,
		m.waitSelection(),
	)
}

// Update updates the parent model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.isInitialized() {
		m.width, m.height = getTermSize()
		return m, tea.Batch(
			func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			},
		)
	}

	var cmds []tea.Cmd

	// Initialize the model if not ready
	if !m.ui.initialized && !m.ui.refreshing {
		m.pipelinesTable.refresh()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ui.width = msg.Width
		m.ui.height = msg.Height

		// update all models/tables
		m.pipelinesTable.Update(msg)
		m.pipelineDetail.Update(msg)
		m.codeBuildDetail.Update(msg)
		m.statusLine.Update(msg)
		m.pager.Update(msg)

		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		log.Debug().Str("model", "tui").Str("func", "Model.Update").Msgf("received KeyMsg: %v", msg)
		activeModel := m.getActiveModel()
		if !m.ui.refreshing {
			switch {
			case key.Matches(msg, allKeys.Quit):
				m.quitting = true
				return m, tea.Quit
			case key.Matches(msg, allKeys.Help):
				m.ui.help = !m.ui.help
				activeModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			case key.Matches(msg, allKeys.PrevTint):
				// TODO: document it in help
				tint.PreviousTint()
				activeModel.Update(redraw{})
			case key.Matches(msg, allKeys.NextTint):
				tint.NextTint()
				activeModel.Update(redraw{})
			}
			if m.ui.inputFocused {
				m.statusLine.Update(msg)
			} else {
				activeModel.Update(msg)
			}
		}

	case tuiMsg:
		log.Debug().Str("model", "tui").Str("func", "Model.Update").Msgf("tuiMsg: %v", msg)
		activeModel := m.getActiveModel()

		switch msg.class {
		case viewUpdate:
			activeModel.Update(msg)

		case viewChange:
			if !supportedView(msg.id) {
				m.statusLineMessage("error", msg.src, fmt.Sprintf("unknown view: %v", msg.class), nil)
				break
			}
			m.ui.views = append(m.ui.views, msg.id)
			m.ui.viewIdx++
			activeModel = m.getActiveModel()
			activeModel.Update(msg)

		case previous:
			m.ui.views = m.ui.views[:m.ui.viewIdx]
			m.ui.viewIdx--
			activeModel = m.getActiveModel()
			activeModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			if msg.trigger {
				// Sleep 1s to allow the job to start
				time.Sleep(1 * time.Second)
				activeModel.Update(refresh{})
			}
		case errorMsg:
			m.statusLineMessage(errorMsg, msg.src, msg.data.(string), msg.reference)

		case input:
			m.statusLineMessage(msg.id, msg.src, msg.data.(string), msg.reference)

		case response:
			activeModel.Update(msg)

		default:
			activeModel = m.getActiveModel()
			activeModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}

		log.Debug().Str("model", "tui").Str("func", "Model.Update").Msgf("views: %v / viewIdx: %v", m.ui.views, m.ui.viewIdx)
		cmds = append(cmds, m.waitSelection())
		return m, tea.Batch(cmds...)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		activeModel := m.getActiveModel()
		activeModel.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// View renders the parent model
func (m *Model) View() string {
	if !m.isInitialized() {
		return ""
	}

	var spinner string
	if m.ui.refreshing {
		spinner = m.spinner.View()
	} else {
		spinner = " "
	}

	activeModel := m.getActiveModel()
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			spinner,
			containerStyle.Render(m.statusLine.View()),
		),
		containerStyle.Render(activeModel.View()),
	)
}

func (m *Model) isInitialized() bool {
	return m.height != 0 && m.width != 0
}

func (m *Model) statusLineMessage(id, src, msg string, ref interface{}) {
	n := notification{
		kind:   id,
		src:    src,
		prompt: msg,
		ref:    ref,
	}
	m.statusLine.Update(n)
}

func (m *Model) getActiveModel() tea.Model {
	switch m.ui.views[m.ui.viewIdx] {
	case "pipelines":
		return m.pipelinesTable
	case "pipeline":
		return m.pipelineDetail
	case "codebuild":
		return m.codeBuildDetail
	case "buildspec", "log":
		return m.pager
	default:
		log.Fatal().Msgf("unknown active model %v", m.ui.views[m.ui.viewIdx])
	}
	return nil
}

func (m *Model) waitSelection() tea.Cmd {
	log.Debug().Str("model", "tui").Str("func", "Model.waitSelection").Msg("function called")
	return func() tea.Msg {
		return tuiMsg(<-m.ui.selection)
	}
}
