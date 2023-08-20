package fullscreen

import (
	lip "github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
)

func (x Model) View() string {
	switch x.mode {
	case modeHelp:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.help.View()),
			x.style.input.Render(x.input.View()))

	case modeInput, modeScroll:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.results.View()),
			x.style.input.Render(x.input.View()))

	case modeQuit:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.quit.View()),
			x.style.input.Render(x.input.View()))

	default:
		panic(errors.Errorf("unexpected mode '%v'", x.mode))
	}
}
