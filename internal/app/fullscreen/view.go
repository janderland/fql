package fullscreen

import (
	lip "github.com/charmbracelet/lipgloss"
)

func (x Model) View() string {
	switch x.mode {
	case modeHelp:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.help.View()),
			x.style.input.Render(x.input.View()))

	default:
		return lip.JoinVertical(lip.Left,
			x.style.results.Render(x.results.View()),
			x.style.input.Render(x.input.View()))
	}
}
