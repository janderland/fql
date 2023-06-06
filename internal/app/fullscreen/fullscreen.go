package fullscreen

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/fullscreen/buffer"
	"github.com/janderland/fdbq/internal/app/fullscreen/results"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/format"
	"github.com/janderland/fdbq/parser/scanner"
)

type App struct {
	Engine engine.Engine
	Format format.Format
	Log    zerolog.Logger
	Out    io.Writer
}

func (x *App) Run(ctx context.Context) error {
	input := textinput.New()
	input.Placeholder = "Query"
	input.Focus()

	model := Model{
		eg:   x.Engine,
		log:  x.Log,
		mode: Input,

		border: Border{
			results: lip.NewStyle().
				Border(lip.RoundedBorder()).
				Padding(0, 1),

			input: lip.NewStyle().
				Border(lip.RoundedBorder()).
				Padding(0, 1),
		},
		results: results.New(x.Format),
		input:   input,
	}

	_, err := tea.NewProgram(
		model,
		tea.WithContext(ctx),
		tea.WithOutput(x.Out),
		tea.WithAltScreen(),
	).Run()
	return err
}

type Model struct {
	eg   engine.Engine
	log  zerolog.Logger
	mode Mode

	border  Border
	results results.Model
	input   textinput.Model
}

type Mode int

const (
	Input Mode = iota
	Scroll
)

type Border struct {
	results lip.Style
	input   lip.Style
}

func (x Model) Init() tea.Cmd {
	return textinput.Blink
}

func (x Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(cursor.BlinkMsg); !ok {
		x.log.Log().Msgf("msg: %T %v", msg, msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return x, tea.Quit

		case tea.KeyRunes:
			if msg.String() == "i" && x.mode == Scroll {
				x.mode = Input
				x.input.Focus()
				return x, textinput.Blink
			}

		case tea.KeyEscape:
			if x.mode == Input {
				x.mode = Scroll
				x.input.Blur()
				return x, nil
			}

		case tea.KeyEnter:
			x.results.Reset()
			return x, doQuery(x.eg, x.input.Value())
		}

	case buffer.StreamBuffer:
		buf, done := msg.Get()
		x.results.PushMany(buf)
		if !done {
			return x, tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
				return msg
			})
		}
		return x, nil

	case error, string, keyval.KeyValue:
		x.results.Push(msg)
		return x, nil

	case tea.WindowSizeMsg:
		return x.updateSize(msg), nil
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			x.results = x.results.Update(msg)
			return x, nil
		}

		switch x.mode {
		case Input:
			x.input, cmd = x.input.Update(msg)
		case Scroll:
			x.results = x.results.Update(msg)
		}
		return x, cmd

	default:
		x.input, cmd = x.input.Update(msg)
		x.results = x.results.Update(msg)
		return x, cmd
	}
}

func (x Model) updateSize(msg tea.WindowSizeMsg) Model {
	const inputLine = 1
	const cursorChar = 1
	inputHeight := x.border.input.GetVerticalFrameSize() + inputLine

	x.border.results.Height(msg.Height - x.border.results.GetVerticalFrameSize() - inputHeight)
	x.border.results.Width(msg.Width - x.border.results.GetHorizontalFrameSize())

	// TODO: I don't know why this +2 is needed.
	x.results.Height(x.border.results.GetHeight() - x.border.results.GetVerticalFrameSize() + 2)

	// TODO: I think -2 is due to a bug with how the textinput bubble renders padding.
	x.input.Width = msg.Width - x.border.input.GetHorizontalFrameSize() - len(x.input.Prompt) - cursorChar - 2
	x.border.input.Width(msg.Width - x.border.input.GetHorizontalFrameSize())

	return x
}

func (x Model) View() string {
	return lip.JoinVertical(lip.Left,
		x.border.results.Render(x.results.View()),
		x.border.input.Render(x.input.View()),
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
			if out == nil {
				return "no results"
			}
			return *out

		case class.ReadRange:
			return buffer.New(eg.ReadRange(context.Background(), kv, engine.RangeOpts{}))

		default:
			return errors.Errorf("unexpected query class '%v'", c)
		}
	}
}
