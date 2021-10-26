package organization

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v39/github"
	te "github.com/muesli/termenv"
)

var (
	Color             = te.ColorProfile().Color
	HasDarkBackground = te.HasDarkBackground()
	Indigo            = NewColorPair("#7571F9", "#5A56E0")
	SubtleIndigo      = NewColorPair("#514DC1", "#7D79F6")
	Cream             = NewColorPair("#FFFDF5", "#FFFDF5")
	YellowGreen       = NewColorPair("#ECFD65", "#04B575")
	Fuschia           = NewColorPair("#EE6FF8", "#EE6FF8")
	Green             = NewColorPair("#04B575", "#04B575")
	Red               = NewColorPair("#ED567A", "#FF4672")
	FaintRed          = NewColorPair("#C74665", "#FF6F91")
	SpinnerColor      = NewColorPair("#747373", "#8E8E8E")
	NoColor           = NewColorPair("", "")
)

type state int
type index int
type OrganizationSetMsg struct {
	Name string
}
type OrganizationErrorMsg struct {
	err error
}

func (e OrganizationErrorMsg) Error() string {
	return e.err.Error()
}

const (
	ready state = iota
	submitting
)

const (
	textInput index = iota
	okButton
	cancelButton
)

type ColorPair struct {
	Dark  string
	Light string
}

type Model struct {
	Done         bool
	Quit         bool
	Organization string
	input        input.Model
	index        index
	errorMessage string
	spinner      spinner.Model
	state        state
	gh           *github.Client
}

func NewColorPair(dark, light string) ColorPair {
	return ColorPair{dark, light}
}

// Color returns the appropriate termenv.Color for the terminal background.
func (c ColorPair) Color() te.Color {
	if HasDarkBackground {
		return Color(c.Dark)
	}

	return Color(c.Light)
}

const prompt = "> "

var focusedPrompt = te.String(prompt).Foreground(Fuschia.Color()).String()

// updateFocus updates the focused states in the model based on the current
// focus index.
func (m *Model) updateFocus() {
	if m.index == textInput && !m.input.Focused() {
		m.input.Focus()
		m.input.Prompt = focusedPrompt
	} else if m.index != textInput && m.input.Focused() {
		m.input.Blur()
		m.input.Prompt = prompt
	}
}

// Move the focus index one unit forward.
func (m *Model) indexForward() {
	m.index++
	if m.index > cancelButton {
		m.index = textInput
	}

	m.updateFocus()
}

// Move the focus index one unit backwards.
func (m *Model) indexBackward() {
	m.index--
	if m.index < textInput {
		m.index = cancelButton
	}

	m.updateFocus()
}

func NewModel(gh *github.Client) Model {
	inputModel := input.NewModel()
	inputModel.Placeholder = "Enter your organization"
	inputModel.Prompt = focusedPrompt
	inputModel.CharLimit = 50
	inputModel.Focus()

	spinnerModel := spinner.NewModel()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style.Foreground(lipgloss.Color(SpinnerColor.Color().Sequence(false)))
	return Model{
		input:        inputModel,
		index:        textInput,
		errorMessage: "",
		spinner:      spinnerModel,
		state:        ready,
		gh:           gh,
	}
}

func Init(gh *github.Client) func() (Model, tea.Cmd) {
	return func() (Model, tea.Cmd) {
		return NewModel(gh), input.Blink
	}
}

func View(m Model) string {
	s := "Enter your Organization\n\n"
	s += m.input.View() + "\n"
	if m.state == submitting {
		s += m.spinner.View() + " Submitting..."
	} else {
		s += OKButtonView(m.index == okButton, true)
		s += CancelButtonView(m.index == cancelButton, false)

		if m.errorMessage != "" {
			s += "\n\n" + te.String(m.errorMessage).Foreground(Red.Color()).String()
		}
	}
	return s
}

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Quit = true
			return m, nil
		case tea.KeyEscape:
			m.Done = true
			return m, nil
		default:
			if m.state == submitting {
				return m, nil
			}

			switch msg.String() {
			case "tab":
				m.indexForward()
			case "shift+tab":
				m.indexBackward()
			case "enter":
				switch m.index {
				case textInput:
					fallthrough
				case okButton:
					m.state = submitting
					m.Organization = strings.TrimSpace(m.input.Value())
					return m, tea.Batch(setOrganization(m), spinner.Tick)
				case cancelButton:
					m.Done = true
					return m, nil
				}
			}

			if m.index == textInput {
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			return m, nil
		}
	case OrganizationErrorMsg:
		m.state = ready
		m.errorMessage = msg.Error()
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

// OKButtonView returns a button reading "OK".
func OKButtonView(focused bool, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("OK", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

// CancelButtonView returns a button reading "Cancel.".
func CancelButtonView(focused bool, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("Cancel", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

func buttonStyling(str string, underline, focused bool) string {
	var s = te.String(str).Foreground(Cream.Color())
	if focused {
		s = s.Background(Fuschia.Color())
	} else {
		s = s.Background(ColorPair{"#827983", "#BDB0BE"}.Color())
	}
	if underline {
		s = s.Underline()
	}

	return s.String()
}

func setOrganization(m Model) tea.Cmd {
	return func() tea.Msg {
		org, _, err := m.gh.Organizations.Get(context.Background(), m.Organization)
		if err != nil {
			return OrganizationErrorMsg{err}
		} else {
			return OrganizationSetMsg{*org.Name}
		}
	}
}
