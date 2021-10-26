package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v39/github"

	"ghtui/ghtui/ui/common"
	"ghtui/ghtui/ui/repositories/repository"
)

type status int

const (
	statusInit status = iota
	statusLoading
	statusReady
	statusRepositorySelected
)

type repositoriesLoadedMsg struct {
	repos []*github.Repository
	items []list.Item
}
type repositorySelectedMsg item
type item struct {
	name        string
	description string
}

type listKeyMap struct {
	toggleHelpMenu   key.Binding
	selectRepository key.Binding
}

type info struct {
	org string
}

type Model struct {
	ctx        string
	list       list.Model
	keys       *listKeyMap
	status     status
	info       info
	gh         *github.Client
	user       *github.User
	spinner    spinner.Model
	repos      []*github.Repository
	repository repository.Model
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.name }

//type delegateKeyMap struct {
//    choose key.Binding
//    down   key.Binding
//}
//
//func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
//    d := list.NewDefaultDelegate()
//
//    d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
//        var title string
//
//        if i, ok := m.SelectedItem().(item); ok {
//            title = i.Title()
//        } else {
//            return nil
//        }
//
//        switch msg := msg.(type) {
//        case tea.KeyMsg:
//            switch {
//            case key.Matches(msg, keys.choose):
//                return m.NewStatusMessage(common.ListStatusMessageStyle().Render("You chose " + title + "."))
//            default:
//                switch {
//                case len(m.Items()) == m.Index()+1:
//                    m.NewStatusMessage(common.ListStatusMessageStyle().Render("You reached the end."))
//                case m.Index() == 0:
//                    m.NewStatusMessage(common.ListStatusMessageStyle().Render("You're at the beginning."))
//                default:
//                    m.NewStatusMessage("")
//                }
//            }
//        }
//
//        return nil
//    }
//
//    help := []key.Binding{keys.choose}
//
//    d.ShortHelpFunc = func() []key.Binding {
//        return help
//    }
//
//    d.FullHelpFunc = func() [][]key.Binding {
//        return [][]key.Binding{help}
//    }
//
//    return d
//}
//
//func newDelegateKeyMap() *delegateKeyMap {
//    return &delegateKeyMap{
//        choose: key.NewBinding(
//            key.WithKeys("enter"),
//            key.WithHelp("enter", "choose"),
//        ),
//        down: key.NewBinding(
//            key.WithKeys("down", "j", "G"),
//            key.WithKeys("down", "j", "G", "down"),
//        ),
//    }
//}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		selectRepository: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select repo"),
		),
	}
}

func NewModel(user *github.User, gh *github.Client) Model {
	return Model{
		gh:      gh,
		user:    user,
		spinner: common.NewSpinnerModel(),
		status:  statusInit,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadRepositories, spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := common.AppStyle().GetPadding()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)
	case tea.KeyMsg:
		if m.status == statusReady {
			switch {
			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil
			case key.Matches(msg, m.keys.selectRepository):
				selectedItem := m.list.VisibleItems()[m.list.Index()].(item)
				return m, common.Cmd(repositorySelectedMsg(selectedItem))
			}
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	case repositoriesLoadedMsg:
		listKeys := newListKeyMap()
		repoList := list.NewModel(msg.items, list.NewDefaultDelegate(), 0, 0)
		repoList.Title = *m.user.Login + " Repositories"
		repoList.Styles.Title = common.ListTitleStyle()
		repoList.ShowPagination()
		topGap, rightGap, bottomGap, leftGap := common.AppStyle().GetPadding()
		width, height := common.ScreenSize()
		repoList.SetSize(width-leftGap-rightGap, height-topGap-bottomGap)
		repoList.AdditionalFullHelpKeys = func() []key.Binding {
			return []key.Binding{
				listKeys.toggleHelpMenu,
			}
		}
		m.repos = msg.repos
		m.keys = listKeys
		m.list = repoList
		m.status = statusReady
		return m, cmd
	case repositorySelectedMsg:
		m.status = statusRepositorySelected
		var repo *github.Repository
		for _, r := range m.repos {
			if *r.Name == msg.name {
				repo = r
				break
			}
		}
		m.repository = repository.NewModel(m.user, repo, m.gh)
		return m, m.repository.Init()
	}

	var cmds []tea.Cmd
	cmds = common.AppendIfNotNil(cmds, cmd)
	m, cmd = updateChildren(m, msg)
	cmds = common.AppendIfNotNil(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func updateChildren(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.status {
	case statusInit:
		m.spinner, cmd = m.spinner.Update(msg)
	case statusReady:
		m.list, cmd = m.list.Update(msg)
		m.list.NewStatusMessage(fmt.Sprintf("Index: %d, Cursor: %d, Visible: %d", m.list.Index(), m.list.Cursor(), len(m.list.VisibleItems())))
	case statusRepositorySelected:
		m.repository, cmd = m.repository.Update(msg)
		if m.repository.Done {
			m.status = statusReady
		}
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.status {
	case statusInit:
		return common.AppStyle().Render(m.spinner.View() + " Loading repositories...")
	case statusReady:
		return common.AppStyle().Render(m.list.View())
	case statusRepositorySelected:
		return m.repository.View()
	}
	return ""
}

func (m Model) loadRepositories() tea.Msg {
	if m.status == statusLoading {
		return spinner.Tick
	}
	m.status = statusLoading
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	var repos []*github.Repository
	var err error
	// gh.Teams.ListTeamReposBySlug(context.Background(), org, team, opts)
	// gh.Repositories.ListByOrg(context.Background(), org, opts)
	repos, _, err = m.gh.Repositories.List(context.Background(), *m.user.Login, &github.RepositoryListOptions{
		ListOptions: *opts,
	})
	if err != nil {
		log.Fatal("Could not get repositories.")
	}

	items := make([]list.Item, len(repos))
	for i := 0; i < len(repos); i++ {
		repo := repos[i]
		description := "The " + *repo.FullName + " repository."
		if repo.Description != nil {
			description = *repo.Description
		}
		items[i] = item{name: *repo.Name, description: description}
	}
	return repositoriesLoadedMsg{repos, items}
}
