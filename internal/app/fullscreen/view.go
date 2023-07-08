package fullscreen

import (
	"regexp"

	lip "github.com/charmbracelet/lipgloss"
)

var (
	helpMsg string
)

func init() {
	const str = `
FDBQ provides an interactive environment for exploring
key-value structures.

The environment has 3 modes: input, scroll, & help. The
environment starts in input mode.

Ctrl+C always quits the program, regardless of the
current mode.

During input mode, the user can type queries into the
input box at the bottom of the screen. Pressing "enter"
cancels the currently executing query, clears the on
screen results, and executes a new query defined by
input box. Pressing "escape" switches to scroll mode.

During scroll mode, the user can scroll through the
results of the previously executed query. Pressing "i"
switches back to input mode. Pressing "?" switches to
help mode.

During help mode, this help screen is displayed.
Pressing "escape" switches to scroll mode.
`

	// Remove lone newlines while leaving blank lines.
	helpMsg = regexp.MustCompile(`([^\n])\n([^\n])`).ReplaceAllString(str, "$1 $2")
}

func (x Model) View() string {
	switch x.mode {
	case modeHelp:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.style.help.Render(helpMsg)),
			x.style.input.Render(x.input.View()))

	default:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.results.View()),
			x.style.input.Render(x.input.View()))
	}
}
