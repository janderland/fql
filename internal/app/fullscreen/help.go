package fullscreen

import (
	"regexp"
	"strings"

	"github.com/janderland/fdbq/internal/app/fullscreen/results"
	"github.com/janderland/fdbq/parser/format"
)

var (
	helpMsg string
)

func init() {
	const str = `
FDBQ provides an interactive environment for interacting
with Foundation DB. The environment has 3 modes: input,
scroll, & help. The environment starts in input mode.
Ctrl+C always quits the program, regardless of the
current mode.

During input mode, the user can type queries into the
input box at the bottom of the screen. Pressing "enter"
cancels the currently executing query, clears the on
screen results, and executes a new query defined by
the input box. Pressing "escape" switches to scroll
mode.

During scroll mode, the user can scroll through the
results of the previously executed query. Pressing "i"
switches back to input mode. Pressing "?" switches to
help mode.

During help mode, this help screen is displayed.
Pressing "escape" switches to scroll mode.
`

	// Remove leading & trailing whitespace.
	helpMsg = strings.TrimSpace(str)

	// Concat paragraphs into single lines.
	helpMsg = regexp.MustCompile(`([^\n])\n([^\n])`).ReplaceAllString(helpMsg, "$1 $2")

	// Remove empty lines.
	helpMsg = regexp.MustCompile(`\n\n([^\n])`).ReplaceAllString(helpMsg, "\n$1")
}

func newHelp() results.Model {
	x := results.New(format.New(format.Cfg{}))
	for _, str := range strings.Split(helpMsg, "\n") {
		x.Push(str)
	}
	return x
}
