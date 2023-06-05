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
		eg:  x.Engine,
		log: x.Log,

		style: Style{
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

type Style struct {
	results lip.Style
	input   lip.Style
}

type Model struct {
	eg  engine.Engine
	log zerolog.Logger

	style   Style
	results results.Model
	input   textinput.Model
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
		const inputLine = 1
		const cursorChar = 1
		inputHeight := x.style.input.GetVerticalFrameSize() + inputLine

		x.style.results.Height(msg.Height - x.style.results.GetVerticalFrameSize() - inputHeight)
		x.style.results.Width(msg.Width - x.style.results.GetHorizontalFrameSize())

		// TODO: I don't know why this +2 is needed.
		x.results.Height(x.style.results.GetHeight() - x.style.results.GetVerticalFrameSize() + 2)

		// TODO: I think -2 is due to a bug with how the textinput bubble renders padding.
		x.input.Width = msg.Width - x.style.input.GetHorizontalFrameSize() - len(x.input.Prompt) - cursorChar - 2
		x.style.input.Width(msg.Width - x.style.input.GetHorizontalFrameSize())
	}

	var cmd tea.Cmd
	x.input, cmd = x.input.Update(msg)
	x.results = x.results.Update(msg)
	return x, cmd
}

func (x Model) View() string {
	return lip.JoinVertical(lip.Left,
		x.style.results.Render(x.results.View()),
		x.style.input.Render(x.input.View()),
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
