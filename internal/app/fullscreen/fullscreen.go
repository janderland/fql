package fullscreen

import (
	"context"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/fullscreen/manager"
	"github.com/janderland/fdbq/internal/app/fullscreen/results"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/format"
)

type App struct {
	Engine engine.Engine
	Format format.Format
	Log    zerolog.Logger
	Out    io.Writer

	Write      bool
	SingleOpts engine.SingleOpts
	RangeOpts  engine.RangeOpts
}

func (x *App) Run(ctx context.Context) error {
	input := textinput.New()
	input.Placeholder = "Query"

	model := Model{
		qm: manager.New(
			ctx,
			x.Engine,
			manager.WithSingleOpts(x.SingleOpts),
			manager.WithRangeOpts(x.RangeOpts),
			manager.WithWrite(x.Write)),

		log:  x.Log,
		mode: modeScroll,

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
	qm     manager.QueryManager
	log    zerolog.Logger
	latest time.Time
	mode   Mode

	border  Border
	results results.Model
	input   textinput.Model
}

type Mode int

const (
	modeScroll Mode = iota
	modeInput
	modeHelp
)

type Border struct {
	results lip.Style
	input   lip.Style
}

func (x Model) Init() tea.Cmd {
	return nil
}

func (x Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(cursor.BlinkMsg); !ok {
		x.log.Log().Msgf("msg: %T %v", msg, msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if model, cmd := x.updateKey(msg); model != nil {
			return *model, cmd
		}

	case manager.AsyncQueryMsg:
		if x.latest.After(msg.StartedAt) {
			return x, nil
		}
		if x.latest.Before(msg.StartedAt) {
			x.results.Reset()
			x.latest = msg.StartedAt
		}

		buf, done := msg.Buffer.Get()
		x.results.PushMany(buf)
		if !done {
			return x, tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
				return msg
			})
		}
		return x, nil

	case error, string, keyval.KeyValue:
		x.results.Reset()
		x.results.Push(msg)
		return x, nil

	case tea.WindowSizeMsg:
		return x.updateSize(msg), nil
	}

	return x.updateChildren(msg)
}

func (x Model) updateKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return &x, tea.Quit
	}

	switch x.mode {
	case modeScroll:
		switch msg.Type {
		case tea.KeyEnter:
			return &x, x.qm.Query(x.input.Value())

		case tea.KeyRunes:
			switch msg.String() {
			case "i":
				x.mode = modeInput
				x.input.Focus()
				return &x, textinput.Blink

			case "?":
				x.mode = modeHelp
				return &x, nil
			}
		}

	case modeInput:
		switch msg.Type {
		case tea.KeyEnter:
			return &x, x.qm.Query(x.input.Value())

		case tea.KeyEscape:
			x.mode = modeScroll
			x.input.Blur()
			return &x, nil
		}

	case modeHelp:
		switch msg.Type {
		case tea.KeyEnter:
			return &x, nil

		case tea.KeyEscape:
			x.mode = modeScroll
			return &x, nil
		}

	default:
		panic(errors.Errorf("unexpected mode '%v'", x.mode))
	}

	return nil, nil
}

func (x Model) updateSize(msg tea.WindowSizeMsg) Model {
	const inputLine = 1
	const cursorChar = 1
	inputHeight := x.border.input.GetVerticalFrameSize() + inputLine

	x.border.results.Height(msg.Height - x.border.results.GetVerticalFrameSize() - inputHeight)
	x.border.results.Width(msg.Width - x.border.results.GetHorizontalFrameSize())
	x.results.Height(x.border.results.GetHeight())

	x.input.Width = msg.Width - x.border.input.GetHorizontalFrameSize() - len(x.input.Prompt) - cursorChar - 2
	x.border.input.Width(msg.Width - x.border.input.GetHorizontalFrameSize())

	return x
}

func (x Model) updateChildren(msg tea.Msg) (tea.Model, tea.Cmd) {
	if x.mode == modeHelp {
		return x, nil
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
		case modeInput:
			x.input, cmd = x.input.Update(msg)
		case modeScroll:
			x.results = x.results.Update(msg)
		}
		return x, cmd

	default:
		x.input, cmd = x.input.Update(msg)
		x.results = x.results.Update(msg)
		return x, cmd
	}
}

// TODO: Include customizable key bindings.
const helpMsg = `
FDBQ provides an interactive environment for exploring
key-value structures.
	
The environment has 3 modes: input, scroll, & help. The
environment starts in input mode.

Ctrl+C always quits the program, regardless of the current
mode.

During input mode, the user can type queries into the input
box at the bottom of the screen. "Enter" cancels the currently
executing query, clears the on screen results, and executes
a new query defined by input box. Pressing "escape" switches to 
scroll mode.

During scroll mode, the user can scroll through the results
of the previously executed query. Pressing "i" switches back
to input mode.

Pressing "Ctrl+?" switches to help mode, regardless of the
current mode. This help screen is displayed during this mode.
Pressing "escape" switches to scroll mode.
`

func (x Model) View() string {
	var upper string
	switch x.mode {
	case modeHelp:
		upper = helpMsg
	default:
		upper = x.results.View()
	}

	return lip.JoinVertical(lip.Left,
		x.border.results.Render(upper),
		x.border.input.Render(x.input.View()))
}
