package common

import "github.com/charmbracelet/lipgloss"

var (
    // https://coolors.co/d88c9a-f2d0a9-f1e3d3-99c1b9-8e7dbe
    danger         = "#d88c9a" // pink
    warning        = "#f2d0a9" // yellow
    white          = "#f1e3d3" // white
    success        = "#99c1b9" // green
    info           = "#8e7dbe" // purple
    gray           = "#505050" // gray
    infoColor      = lipgloss.AdaptiveColor{Light: info, Dark: info}
    successColor   = lipgloss.AdaptiveColor{Light: success, Dark: success}
    dangerColor    = lipgloss.AdaptiveColor{Light: danger, Dark: danger}
    grayColor      = lipgloss.AdaptiveColor{Light: gray, Dark: gray}
    whiteColor     = lipgloss.AdaptiveColor{Light: white, Dark: white}
    appStyle       = lipgloss.NewStyle().Padding(2, 2)
    errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(danger))
    listTitleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(white)).
        Background(lipgloss.Color(success)).
        Padding(0, 1)
    listStatusMessageStyle = lipgloss.NewStyle().
        Foreground(successColor)
)

func DangerColor() lipgloss.AdaptiveColor {
    return dangerColor
}

func WhiteColor() lipgloss.AdaptiveColor {
    return whiteColor
}

func AppStyle() lipgloss.Style {
    return appStyle
}

func ErrorStyle() lipgloss.Style {
    return errorStyle
}

func ListTitleStyle() lipgloss.Style {
    return listTitleStyle
}

func ListStatusMessageStyle() lipgloss.Style {
    return listStatusMessageStyle
}

func PaneActiveStyle(width int, height int) lipgloss.Style {
    return paneStyle(successColor, width, height)
}

func PaneInactiveStyle(width int, height int) lipgloss.Style {
    return paneStyle(grayColor, width, height)
}

func PaneSelectedItemStyle() lipgloss.Style {
    return lipgloss.NewStyle().Foreground(whiteColor).Background(infoColor)
}

func paneStyle(color lipgloss.AdaptiveColor, width int, height int) lipgloss.Style {
    return lipgloss.NewStyle().
        BorderForeground(color).
        Border(lipgloss.NormalBorder()).
        Width(width).
        Height(height)
}
