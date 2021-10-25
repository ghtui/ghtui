package common

import tea "github.com/charmbracelet/bubbletea"

func AppendIfNotNil(cmds []tea.Cmd, cmd tea.Cmd) []tea.Cmd {
    if cmd != nil {
       cmds = append(cmds, cmd)
    }
    return cmds
}
