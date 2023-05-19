package main

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/engine/stream"
	"github.com/janderland/fdbq/internal/app/fullscreen/buffer"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/format"
	"github.com/janderland/fdbq/parser/scanner"
)

func main() {
	if err := fdb.APIVersion(620); err != nil {
		panic(err)
	}

	db, err := fdb.OpenDefault()
	if err != nil {
		panic(err)
	}

	eg := engine.New(facade.NewTransactor(db, directory.Root()))

	file, err := tea.LogToFile("log.txt", "fdbq")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(errors.Wrap(err, "failed to close log file"))
		}
	}()

	p := tea.NewProgram(newModel(eg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type Style struct {
	results lip.Style
	input   lip.Style
}

type Model struct {
	eg engine.Engine

	list  *list.List
	lines []string

	style Style
	input textinput.Model
}

func newModel(eg engine.Engine) Model {
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
		eg:    eg,
		list:  list.New(),
		input: input,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("msg: %T %v", msg, msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			m.list = list.New()
			return m, doQuery(m.eg, m.input.Value())
		}

	case buffer.StreamBuffer:
		buf, done := msg.Get()
		for item := buf.Front(); item != nil; item = item.Next() {
			m.list.PushFront(item.Value)
		}
		if !done {
			return m, tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
				return msg
			})
		}
		return m, nil

	case keyval.KeyValue, error:
		m.list.PushFront(msg)
		return m, nil

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
	f := format.New(format.Cfg{})
	item := m.list.Front()
	i := -1

	for i = range m.lines {
		if item == nil {
			break
		}

		switch val := item.Value.(type) {
		case keyval.KeyValue:
			f.Reset()
			f.KeyValue(val)
			m.lines[i] = f.String()

		case stream.KeyValErr:
			if val.Err != nil {
				m.lines[i] = fmt.Sprintf("ERR! %v", val)
			} else {
				f.Reset()
				f.KeyValue(val.KV)
				m.lines[i] = f.String()
			}

		case stream.DirErr:
			if val.Err != nil {
				m.lines[i] = fmt.Sprintf("ERR! %v", val)
			} else {
				f.Reset()
				f.Directory(convert.FromStringArray(val.Dir.GetPath()))
				m.lines[i] = f.String()
			}

		case string:
			m.lines[i] = fmt.Sprintf("# %s", val)

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

func doQuery(eg engine.Engine, str string) func() tea.Msg {
	return func() tea.Msg {
		p := parser.New(scanner.New(strings.NewReader(str)))
		query, err := p.Parse()
		if err != nil {
			return err
		}

		if query, ok := query.(keyval.Directory); ok {
			return buffer.New(eg.Directories(context.Background(), query))
		}

		var kv keyval.KeyValue
		if key, ok := query.(keyval.Key); ok {
			kv = keyval.KeyValue{Key: key, Value: keyval.Variable{}}
		} else {
			kv = query.(keyval.KeyValue)
		}

		switch c := class.Classify(kv); c {
		case class.Constant:
			if err := eg.Set(kv); err != nil {
				return err
			}
			return "key set"

		case class.Clear:
			if err := eg.Clear(kv); err != nil {
				return err
			}
			return "key cleared"

		case class.ReadSingle:
			out, err := eg.ReadSingle(kv, engine.SingleOpts{})
			if err != nil {
				return err
			}
			return *out

		case class.ReadRange:
			return buffer.New(eg.ReadRange(context.Background(), kv, engine.RangeOpts{}))

		default:
			return errors.Errorf("unexpected query class '%v'", c)
		}
	}
}
