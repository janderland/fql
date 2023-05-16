package main

import (
	"container/list"
	"fmt"
	"log"
	"math/rand"
	"strings"

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
	list  list.List
	lines []string
	count int

	style Style
	input textinput.Model
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

		case tea.KeyEnter:
			if rand.Float32() > 0.5 {
				m.list.PushFront(fmt.Sprintf("/my/dir{%d, %f}=nil", m.count, rand.Float32()))
			} else {
				m.list.PushFront(fmt.Errorf("this is a failure: %d things wrong", m.count))
			}
			m.count++
		}

	case tea.WindowSizeMsg:
		const inputLine = 1
		const cursorChar = 1
		inputHeight := m.style.input.GetVerticalFrameSize() + inputLine

		m.style.results.Height(msg.Height - m.style.results.GetVerticalFrameSize() - inputHeight)
		m.style.results.Width(msg.Width - m.style.results.GetHorizontalFrameSize())
		m.lines = make([]string, m.style.results.GetHeight())

		// I think -2 is due to a bug with how the textinput bubble renders padding.
		m.input.Width = msg.Width - m.style.input.GetHorizontalFrameSize() - len(m.input.Prompt) - cursorChar - 2
		m.style.input.Width(msg.Width - m.style.input.GetHorizontalFrameSize())
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	i := -1
	item := m.list.Front()
	for i = range m.lines {
		if item == nil {
			break
		}

		switch val := item.Value.(type) {
		case string:
			m.lines[i] = val
		case error:
			m.lines[i] = fmt.Sprintf("ERR! %v", val)
		default:
			m.lines[i] = fmt.Sprintf("ERR! unexpected item value '%T'", val)
		}

		item = item.Next()
	}

	var results strings.Builder
	if i >= 0 {
		for j := i; j >= 0; j-- {
			results.WriteString(m.lines[j])
			results.WriteRune('\n')
		}
	}

	return lip.JoinVertical(lip.Left,
		m.style.results.Render(results.String()),
		m.style.input.Render(m.input.View()),
	)
}
