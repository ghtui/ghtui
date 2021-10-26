package common

import (
	"bytes"
	"log"
	"strings"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func AppendIfNotNil(cmds []tea.Cmd, cmd tea.Cmd) []tea.Cmd {
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func ScreenSize() (int, int) {
	width, height, _ := term.GetSize(0)
	return width, height
}

func Highlight(name string, contents string) string {
	lexer := lexers.Match(name)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	var buff bytes.Buffer
	var filetype string
	if strings.Index(name, ".") > 0 {
		filetype = name[strings.LastIndex(name, "."):len(name)]
	} else {
		filetype = name
	}
	err := quick.Highlight(&buff, contents, filetype, "terminal16m", "monokai")
	if err != nil {
		log.Fatal(err)
		return contents
	}
	return buff.String()
}
