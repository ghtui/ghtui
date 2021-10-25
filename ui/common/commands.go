package common

import tea "github.com/charmbracelet/bubbletea"

func BatchCommands(cmds... tea.Cmd) tea.Cmd {
    var output []tea.Cmd
    for _, cmd := range cmds {
        if cmd != nil {
            output = append(output, cmd)
        }
    }
    return tea.Batch(output...)
}
