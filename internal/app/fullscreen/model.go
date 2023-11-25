package fullscreen

import (
	"context"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/fullscreen/manager"
	"github.com/janderland/fdbq/internal/app/fullscreen/results"
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
		mode: modeScroll,
		log:  x.Log,

		style: Style{
			results: lip.NewStyle().
				Padding(0, 1),

			input: lip.NewStyle().
				Border(lip.RoundedBorder()).
				Padding(0, 1),
		},

		results: results.New(
			results.WithFormat(x.Format),
			results.WithLogger(x.Log)),

		help:  newHelp(),
		quit:  newQuit(),
		input: input,

		qm: manager.New(
			ctx,
			x.Engine,
			manager.WithSingleOpts(x.SingleOpts),
			manager.WithRangeOpts(x.RangeOpts),
			manager.WithWrite(x.Write)),
	}

	_, err := tea.NewProgram(
		model,
		tea.WithContext(ctx),
		tea.WithOutput(x.Out),
		tea.WithAltScreen(),
	).Run()
	return err
}

type Mode int

const (
	modeScroll Mode = iota
	modeInput
	modeHelp
	modeQuit
)

type Style struct {
	results lip.Style
	input   lip.Style
}

type Model struct {
	mode   Mode
	latest time.Time
	log    zerolog.Logger

	style   Style
	results results.Model
	help    results.Model
	quit    results.Model
	input   textinput.Model
	qm      manager.QueryManager
}
