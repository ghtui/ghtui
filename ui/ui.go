package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v39/github"

	"ghtui/ghtui/ui/activity"
	"ghtui/ghtui/ui/common"
	"ghtui/ghtui/ui/repositories"
)

type status int

const (
	statusInit status = iota
	statusLoading
	statusReady
)

type model struct {
	quit         bool
	done         bool
	username     string
	gh           *github.Client
	spinner      spinner.Model
	status       status
	activity     activity.Model
	repositories repositories.Model
	errorMsg     string
	user         *github.User
}

type userLoadedMsg *github.User
type errorMsg error

func NewProgram(username string, gh *github.Client) *tea.Program {
	return tea.NewProgram(initialModel(username, gh), tea.WithAltScreen())
}

func initialModel(username string, gh *github.Client) model {
	return model{
		username: username,
		status:   statusInit,
		gh:       gh,
		spinner:  common.NewSpinnerModel(),
	}
}

func (m model) Init() tea.Cmd {
	return spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quit = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// TODO implement window resizing?
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	case userLoadedMsg:
		m.user = msg
		m.username = *msg.Login
		m.status = statusReady
		m.repositories = repositories.NewModel(msg, m.gh)
		cmd = m.repositories.Init()
	case errorMsg:
		m.errorMsg = msg.Error()
	}

	var cmds []tea.Cmd
	cmds = common.AppendIfNotNil(cmds, cmd)
	m, cmd = updateChildren(m, msg)
	cmds = common.AppendIfNotNil(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func updateChildren(m model, msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.status {
	case statusInit:
		m.spinner, cmd = m.spinner.Update(msg)
		cmd = tea.Batch(cmd, m.loadUserCmd)
		m.status = statusLoading
	case statusLoading:
		m.spinner, cmd = m.spinner.Update(msg)
	case statusReady:
		m.repositories, cmd = m.repositories.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	s := ""
	switch m.status {
	case statusInit:
	case statusLoading:
		s += common.AppStyle().Render(m.spinner.View() + " Loading user...")
	case statusReady:
		s += m.repositories.View()
	}

	if m.errorMsg != "" {
		s += "\n\n" + m.errorMsg
	}

	return lipgloss.JoinVertical(lipgloss.Top, s)
}

func (m model) loadUserCmd() tea.Msg {
	if m.status == statusLoading {
		return spinner.Tick()
	}

	user, _, err := m.gh.Users.Get(context.Background(), m.username)
	if err != nil {
		return errorMsg(err)
	}
	return userLoadedMsg(user)
}
