package repository

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v39/github"

	"ghtui/ghtui/ui/common"
	"ghtui/ghtui/ui/repositories/repository/pane"
)

type repositoryFilesLoadedMsg []*github.RepositoryContent
type repositoryFileLoadedMsg *github.RepositoryContent
type repositoryErrorMsg error
type status int

const (
	statusInit status = iota
	statusLoading
	statusReady
)

type Model struct {
	Done bool
	Quit bool

	user             *github.User
	repository       *github.Repository
	spinner          spinner.Model
	status           status
	paneIndex        int
	fileIndex        int
	gh               *github.Client
	statusMsg        string
	contents         []*github.RepositoryContent
	selectedContents *github.RepositoryContent
	path             string
	leftPane         pane.Model
	rightPane        pane.Model
	title            string
	depth            int
}

func NewModel(user *github.User, repository *github.Repository, gh *github.Client) Model {
	return Model{
		Done:       false,
		Quit:       false,
		user:       user,
		repository: repository,
		spinner:    common.NewSpinnerModel(),
		status:     statusInit,
		paneIndex:  0,
		fileIndex:  0,
		gh:         gh,
		path:       "",
		depth:      0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadRepositoryContents, spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// TODO: do something?
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if m.paneIndex == 0 {
				if m.depth == 0 {
					m.Done = true
					return m, nil
				} else {
					m.depth -= 1
					m.path = m.path[0:strings.LastIndex(m.path, "/")]
					return m, common.Cmd(m.loadRepositoryContents())
				}
			} else if m.paneIndex == 1 {
				m.selectedContents = nil
				m.paneIndex = 0
			}
		default:
			switch msg.String() {
			case "enter":
				return loadRepositoryContent(m)
			case "down":
				if m.paneIndex == 0 {
					if m.fileIndex < len(m.contents)-1 {
						m.fileIndex += 1
					}
				} else if m.paneIndex == 1 {
					m.rightPane.Viewport.LineDown(1)
				}
			case "up":
				if m.paneIndex == 0 {
					if m.fileIndex > 0 {
						m.fileIndex -= 1
					}
				} else if m.paneIndex == 1 {
					m.rightPane.Viewport.LineUp(1)
				}
			case "tab":
				// Switch between 0 and 1
				m.paneIndex ^= 1
			}
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	case repositoryFilesLoadedMsg:
		m.fileIndex = 0
		m.status = statusReady
		m.contents = msg
		titleStyle := common.ListTitleStyle()
		m.title = titleStyle.Render(*m.repository.Name)
		width, height := common.ScreenSize()
		baseWidth := width / 4
		top, right, bottom, _ := common.AppStyle().GetPadding()
		paneHeight := height - top - bottom
		m.leftPane = pane.NewModel(baseWidth-right, paneHeight-3, m.paneIndex == 0)
		m.rightPane = pane.NewModel(baseWidth*3-right, paneHeight-3, m.paneIndex == 1)
		m.leftPane.Viewport.SetContent(m.getFileList())
		m.rightPane.Viewport.SetContent("Use the arrow keys to navigate. Press enter to select a file/folder.")
	case repositoryFileLoadedMsg:
		m.status = statusReady
		m.paneIndex = 1
		m.selectedContents = msg
		bytes, _ := base64.StdEncoding.DecodeString(*m.selectedContents.Content)
		m.rightPane.Viewport.SetContent(common.Highlight(*m.selectedContents.Name, string(bytes)))
	case repositoryErrorMsg:
		m.status = statusReady
		m.statusMsg = msg.Error()
	}
	var childCmd tea.Cmd
	m, childCmd = updateChildren(m, msg)
	common.BatchCommands(cmd, childCmd)
	return m, cmd
}

func updateChildren(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.status {
	case statusInit:
	case statusLoading:
	case statusReady:
		m.leftPane.Active = m.paneIndex == 0
		m.rightPane.Active = m.paneIndex == 1
		// do nothing?
	}
	return m, cmd
}

func (m Model) View() string {
	s := ""
	switch m.status {
	case statusInit:
		fallthrough
	case statusLoading:
		s += m.spinner.View() + " Loading " + *m.repository.Name + "..."
	case statusReady:
		m.leftPane.Viewport.SetContent(m.getFileList())
		panes := lipgloss.JoinHorizontal(lipgloss.Top, m.leftPane.View(), m.rightPane.View())
		var statusBar string
		if m.statusMsg != "" {
			statusBar = m.spinner.View() + " " + m.statusMsg
		}
		s += lipgloss.JoinVertical(lipgloss.Top, m.title, panes, statusBar)
	}
	s += string(rune(m.status))
	return s
}

func (m Model) loadRepositoryContents() tea.Msg {
	if m.status == statusLoading {
		return spinner.Tick
	}

	m.status = statusLoading
	m.statusMsg = "Loading repository contents..."
	opts := &github.RepositoryContentGetOptions{Ref: *m.repository.DefaultBranch}
	_, directory, _, err := m.gh.Repositories.GetContents(context.Background(), *m.user.Login, *m.repository.Name, m.path, opts)
	if err != nil {
		return repositoryErrorMsg(err)
	}

	return repositoryFilesLoadedMsg(directory)
}

func loadRepositoryContent(m Model) (Model, tea.Cmd) {
	contents := m.contents[m.fileIndex]
	m.statusMsg = "Loading " + *contents.Name + " " + *contents.Type
	if *contents.Type == "dir" {
		m.path = m.path + "/" + *contents.Name
		m.depth += 1
		return m, common.Cmd(m.loadRepositoryContents())
	} else if *contents.Type == "file" {
		m.statusMsg = "Loading " + *contents.Name + "..."
		file, _, _, err := m.gh.Repositories.GetContents(
			context.Background(),
			*m.user.Login,
			*m.repository.Name,
			m.path+"/"+*contents.Name,
			&github.RepositoryContentGetOptions{
				Ref: *m.repository.DefaultBranch,
			},
		)
		if err != nil {
			return m, common.Cmd(repositoryErrorMsg(err))
		}
		return m, common.Cmd(repositoryFileLoadedMsg(file))
	} else {
		// This should never happen.
		return m, nil
	}
}

func (m Model) getFileList() string {
	pane := ""
	for i, content := range m.contents {
		line := ""
		if *content.Type == "file" {
			line += "📄"
		} else if *content.Type == "dir" {
			line += "📁"
		}
		line += " " + *content.Name
		if i == m.fileIndex {
			pane += common.PaneSelectedItemStyle().Render(line) + "\n"
		} else {
			pane += line + "\n"
		}
	}
	return pane
}
