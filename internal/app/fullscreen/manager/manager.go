package manager

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/fullscreen/buffer"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/scanner"
)

type AsyncQueryMsg struct {
	StartedAt time.Time
	Buffer    buffer.StreamBuffer
}

type Option func(*QueryManager)

type QueryManager struct {
	eg         engine.Engine
	singleOpts engine.SingleOpts
	rangeOpts  engine.RangeOpts
	write      bool

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, eg engine.Engine, opts ...Option) QueryManager {
	x := QueryManager{
		eg:     eg,
		ctx:    ctx,
		cancel: func() {},
	}
	for _, option := range opts {
		option(&x)
	}
	return x
}

func WithSingleOpts(opts engine.SingleOpts) Option {
	return func(x *QueryManager) {
		x.singleOpts = opts
	}
}

func WithRangeOpts(opts engine.RangeOpts) Option {
	return func(x *QueryManager) {
		x.rangeOpts = opts
	}
}

func WithWrite(write bool) Option {
	return func(x *QueryManager) {
		x.write = write
	}
}

func (x *QueryManager) Query(str string) func() tea.Msg {
	// Cancel previous query before starting a new one.
	x.cancel()

	// Create a new context for the new query.
	var childCtx context.Context
	childCtx, x.cancel = context.WithCancel(x.ctx)

	return func() tea.Msg {
		p := parser.New(scanner.New(strings.NewReader(str)))
		query, err := p.Parse()
		if err != nil {
			return err
		}

		if query, ok := query.(keyval.Directory); ok {
			return AsyncQueryMsg{
				StartedAt: time.Now(),
				Buffer:    buffer.New(x.eg.Directories(childCtx, query)),
			}
		}

		var kv keyval.KeyValue
		if key, ok := query.(keyval.Key); ok {
			kv = keyval.KeyValue{Key: key, Value: keyval.Variable{}}
		} else {
			kv = query.(keyval.KeyValue)
		}

		switch c := class.Classify(kv); c {
		case class.Constant:
			if err := x.eg.Set(kv); err != nil {
				return err
			}
			return "key set"

		case class.Clear:
			if err := x.eg.Clear(kv); err != nil {
				return err
			}
			return "key cleared"

		case class.ReadSingle:
			out, err := x.eg.ReadSingle(kv, x.singleOpts)
			if err != nil {
				return err
			}
			if out == nil {
				return "no results"
			}
			return *out

		case class.ReadRange:
			return AsyncQueryMsg{
				StartedAt: time.Now(),
				Buffer:    buffer.New(x.eg.ReadRange(childCtx, kv, x.rangeOpts)),
			}

		default:
			return errors.Errorf("unexpected query class '%v'", c)
		}
	}
}
