package manager

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/fullscreen/buffer"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/scanner"
)

type QueryManager struct {
	eg         engine.Engine
	singleOpts engine.SingleOpts
	rangeOpts  engine.RangeOpts

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, eg engine.Engine, singleOpts engine.SingleOpts, rangeOpts engine.RangeOpts) QueryManager {
	return QueryManager{
		eg:         eg,
		singleOpts: singleOpts,
		rangeOpts:  rangeOpts,

		ctx:    ctx,
		cancel: func() {},
	}
}

func (x *QueryManager) Query(str string) func() tea.Msg {
	return func() tea.Msg {
		p := parser.New(scanner.New(strings.NewReader(str)))
		query, err := p.Parse()
		if err != nil {
			return err
		}

		if query, ok := query.(keyval.Directory); ok {
			return buffer.New(x.eg.Directories(x.newChildCtx(), query))
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
			return buffer.New(x.eg.ReadRange(x.newChildCtx(), kv, x.rangeOpts))

		default:
			return errors.Errorf("unexpected query class '%v'", c)
		}
	}
}

func (x *QueryManager) newChildCtx() context.Context {
	// Cancel the old child before creating a new one.
	x.cancel()

	var ctx context.Context
	ctx, x.cancel = context.WithCancel(x.ctx)
	return ctx
}
