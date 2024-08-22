package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.NormalBorder()

		return lipgloss.NewStyle().BorderStyle(b).BorderForeground(lipgloss.Color("240")).
			BorderTop(true).BorderBottom(true).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	}()
)

// PagerSelector is the data for the pager
type PagerSelector struct {
	name    string
	token   *string
	content string
}

// Pager represent a pager
type Pager struct {
	*viewport.Model
	name         string
	pathTitle    string
	title        string
	content      string
	lastLogToken *string
	msg          *tuiMsg
	ui           *uiData
	help         help.Model
}

// NewPager returns a new Pager
func NewPager(ui *uiData) *Pager {
	p := viewport.New(1, 1)

	return &Pager{
		Model: &p,
		ui:    ui,
		help:  help.New(),
	}
}

// Init implement the tea.Model interface
func (m *Pager) Init() tea.Cmd {
	return nil
}

// Update implement the tea.Model interface
func (m *Pager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.ui.updatPath(m.pathTitle)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight + 1

		if m.ui.help {
			verticalMarginHeight += helpFullHeiggt
		} else {
			verticalMarginHeight += helpHeight
		}
		m.Width = msg.Width
		m.Height = msg.Height - verticalMarginHeight

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, allKeys.Previous):
			m.ui.previousView()
		case key.Matches(msg, allKeys.Refresh):
			m.refreshLog()
		}

	case tuiMsg:
		log.Debug().Str("model", "tui").Str("func", "Pager.Update").Msgf("tuiMsg class: %v, id: %v, trigger: %v, data: %v", msg.class, msg.id, msg.trigger, msg.data)
		m.msg = &msg

		switch msg.class {
		case viewChange:
			m.name = m.msg.data.(PagerSelector).name
			m.SetContent()
		case viewUpdate:
			m.content += msg.data.(PagerSelector).content
			m.lastLogToken = msg.data.(PagerSelector).token
			m.Model.SetContent(m.content)
		}
	}
	*m.Model, _ = m.Model.Update(msg)
	return m, nil
}

// View implement the tea.Model interface
func (m *Pager) View() string {
	var help string
	if m.ui.help {
		help = m.helpViewFull()
	} else {
		help = m.helpView()
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.Model.View(), m.footerView(), help)
}

// SetContent set the content of the pager
func (m *Pager) SetContent() {
	m.reset()
	switch m.msg.id {
	case buildspecView:
		m.title = "CodeBuild Buildspec Definition"
		m.pathTitle = "buildspec"

		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.Width),
		)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create renderer")
		}

		bs := m.ui.dataCache.codebuilds[m.name].Project.Projects[0].Source.Buildspec
		m.content, err = renderer.Render(fmt.Sprintf("```yaml\n%v\n```", *bs))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to render markdown")
		}
		m.Model.SetContent(m.content)

	case logView:
		m.title = "CodeBuild Exection Log " + m.name
		m.pathTitle = "cloudwatch-logs"
		m.reset()
		m.refreshLog()
	}
}

func (m *Pager) headerView() string {
	title := titleStyle.Render(m.title + strings.Repeat(" ", m.Width-lipgloss.Width(m.title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title)
}

func (m *Pager) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.ScrollPercent()*100))
	bottomLineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	line := strings.Repeat(bottomLineStyle.Render("─"), max(0, m.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m *Pager) reset() {
	var token *string
	m.lastLogToken = token
	m.content = ""
	m.Model.SetContent(m.content)
}

func (m *Pager) helpView() string {
	return m.help.ShortHelpView([]key.Binding{
		allKeys.Up,
		allKeys.Down,
		allKeys.Previous,
		allKeys.Refresh,
		allKeys.Quit,
		allKeys.Help,
	})
}

func (m *Pager) helpViewFull() string {
	return m.help.FullHelpView([][]key.Binding{
		{
			allKeys.Up,
			allKeys.Down,
			allKeys.Previous,
		},
		{
			allKeys.Refresh,
			allKeys.Quit,
			allKeys.Help,
		},
	})
}
