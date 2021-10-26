package activity

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v39/github"
)

type Model struct {
	list     list.Model
	username string
	errMsg   string
}

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

func NewModel(username string, gh *github.Client) Model {
	opts := &github.ListOptions{
		Page:    0,
		PerPage: 20,
	}
	events, _, err := gh.Activity.ListEventsPerformedByUser(context.Background(), username, false, opts)
	if err != nil {
		return Model{
			errMsg: err.Error(),
		}
	}
	items := make([]list.Item, len(events))
	for i := 0; i < len(events); i++ {
		//payload, _ := event.ParsePayload()
		//createEvent := payload.(github.CreateEvent)
		event := events[i]
		items[i] = item{
			title:       *event.Actor.Login + " performed " + *event.Type + " on " + *event.Repo.Name,
			description: "Performed at " + event.CreatedAt.String(),
		}
	}
	eventList := list.NewModel(items, list.NewDefaultDelegate(), 0, 0)
	eventList.Title = username + " Events"
	eventList.Styles.Title = titleStyle

	return Model{
		list:     eventList,
		username: username,
		errMsg:   "",
	}
}

func (m Model) Init() tea.Cmd {
	return input.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)
	}
	newModel, cmd := m.list.Update(msg)
	m.list = newModel
	return m, cmd
}

func (m Model) View() string {
	return appStyle.Render(m.list.View())
}
