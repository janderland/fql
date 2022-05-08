package headless

import (
	"context"

	"github.com/janderland/fdbq/engine"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/pkg/errors"
)

type op struct {
	ctx context.Context
	app *App
	eg  engine.Engine
	err error
}

var _ q.QueryOperation = &op{}

func newOp(ctx context.Context, app *App, eg engine.Engine) op {
	return op{ctx: ctx, app: app, eg: eg}
}

func (x *op) Do(query q.Query) error {
	query.Query(x)
	return x.err
}

func (x *op) ForDirectory(query q.Directory) {
	if err := x.app.directories(x.ctx, x.eg, query); err != nil {
		x.err = errors.Wrap(err, "failed to execute as directory query")
	}
}

func (x *op) ForKey(query q.Key) {
	if err := x.app.singleRead(
		x.eg,
		q.KeyValue{
			Key:   query,
			Value: q.Variable{},
		},
	); err != nil {
		x.err = errors.Wrap(err, "failed to execute as single read query")
	}
}

func (x *op) ForKeyValue(query q.KeyValue) {
	switch c := class.Classify(query); c {
	case class.Constant:
		if err := x.app.set(x.eg, query); err != nil {
			x.err = errors.Wrap(err, "failed to execute as set query")
		}

	case class.Clear:
		if err := x.app.clear(x.eg, query); err != nil {
			x.err = errors.Wrap(err, "failed to execute as clear query")
		}

	case class.SingleRead:
		if err := x.app.singleRead(x.eg, query); err != nil {
			x.err = errors.Wrap(err, "failed to execute as single read query")
		}

	case class.RangeRead:
		if err := x.app.rangeRead(x.ctx, x.eg, query); err != nil {
			x.err = errors.Wrap(err, "failed to execute as range read query")
		}

	default:
		x.err = errors.Errorf("unexpected query class '%v'", c)
	}
}
