package common

import "github.com/charmbracelet/bubbles/spinner"

func NewSpinnerModel() spinner.Model {
	return spinner.Model{Spinner: spinner.MiniDot}
}
