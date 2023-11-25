package fullscreen

import (
	lip "github.com/charmbracelet/lipgloss"
)

func (x Model) View() string {
	return lip.JoinVertical(lip.Left,
		x.style.results.Render(x.results.Top().View()),
		x.style.input.Render(x.input.View()))
}
