package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type Style struct {
	results lip.Style
	input   lip.Style
}

type Model struct {
	style Style

	input textinput.Model
	err   error
}

func newModel() Model {
	input := textinput.New()
	input.Placeholder = "Query"
	input.Focus()

	return Model{
		style: Style{
			results: lip.NewStyle().
				Border(lip.RoundedBorder()).
				Padding(0, 1),

			input: lip.NewStyle().
				Border(lip.RoundedBorder()).
				Padding(0, 1),
		},
		input: input,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		const inputLine = 1
		const cursorChar = 1
		inputHeight := m.style.input.GetVerticalFrameSize() + inputLine

		m.style.results.Height(msg.Height - m.style.results.GetVerticalFrameSize() - inputHeight)
		m.style.results.Width(msg.Width - m.style.results.GetHorizontalFrameSize())

		// I think -2 is due to a bug with how the textinput bubble renders padding.
		m.input.Width = msg.Width - m.style.input.GetHorizontalFrameSize() - len(m.input.Prompt) - cursorChar - 2
		m.style.input.Width(msg.Width - m.style.input.GetHorizontalFrameSize())
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lip.JoinVertical(lip.Left,
		m.style.results.Render(""),
		m.style.input.Render(m.input.View()),
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
