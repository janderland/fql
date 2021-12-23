package headless

import (
	"context"
	"fmt"
	"io"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type App struct {
	Flags flag.Flags
	Log   zerolog.Logger
	Out   io.Writer
}

func (x *App) Run(ctx context.Context, db facade.Transactor, queries []string) error {
	eg := engine.New(ctx, db)
	_, err := eg.Transact(func(eg engine.Engine) (interface{}, error) {
		for _, str := range queries {
			if err := x.query(eg, str); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (x *App) query(eg engine.Engine, str string) error {
	query, onlyDir, err := parser.ParseQuery(str)
	if err != nil {
		return errors.Wrap(err, "failed to parse query")
	}

	if onlyDir {
		if err := x.directories(eg, query.Key.Directory); err != nil {
			return errors.Wrap(err, "failed to execute as directory query")
		}
		return nil
	}

	switch c := class.Classify(*query); c {
	case class.Constant:
		if err := x.set(eg, *query); err != nil {
			return errors.Wrap(err, "failed to execute as set query")
		}
		return nil

	case class.Clear:
		if err := x.clear(eg, *query); err != nil {
			return errors.Wrap(err, "failed to execute as clear query")
		}
		return nil

	case class.SingleRead:
		if err := x.singleRead(eg, *query); err != nil {
			return errors.Wrap(err, "failed to execute as single read query")
		}
		return nil

	case class.RangeRead:
		if err := x.rangeRead(eg, *query); err != nil {
			return errors.Wrap(err, "failed to execute as range read query")
		}
		return nil

	default:
		return errors.Errorf("unexpected query class '%v'", c)
	}
}

func (x *App) set(eg engine.Engine, query q.KeyValue) error {
	if !x.Flags.Write {
		return errors.New("writing isn't enabled")
	}
	x.Log.Log().Interface("query", query).Msg("executing set query")
	return eg.Set(query, x.Flags.ByteOrder())
}

func (x *App) clear(eg engine.Engine, query q.KeyValue) error {
	if !x.Flags.Write {
		return errors.New("writing isn't enabled")
	}
	x.Log.Log().Interface("query", query).Msg("executing clear query")
	return eg.Clear(query)
}

func (x *App) singleRead(eg engine.Engine, query q.KeyValue) error {
	x.Log.Log().Interface("query", query).Msg("executing single-read query")
	kv, err := eg.SingleRead(query, x.Flags.SingleOpts())
	if err != nil {
		return err
	}
	if kv == nil {
		return nil
	}
	str, err := parser.FormatKeyValue(*kv)
	if err != nil {
		return errors.Wrap(err, "failed to format output")
	}
	if _, err := fmt.Fprintln(x.Out, str); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (x *App) rangeRead(eg engine.Engine, query q.KeyValue) error {
	x.Log.Log().Interface("query", query).Msg("executing range-read query")
	for kv := range eg.RangeRead(context.Background(), query, x.Flags.RangeOpts()) {
		if kv.Err != nil {
			return kv.Err
		}
		str, err := parser.FormatKeyValue(kv.KV)
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}
		if _, err := fmt.Fprintln(x.Out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}

func (x *App) directories(eg engine.Engine, query q.Directory) error {
	x.Log.Log().Interface("query", query).Msg("executing directory query")
	for dir := range eg.Directories(context.Background(), query) {
		if dir.Err != nil {
			return dir.Err
		}
		str, err := parser.FormatDirectory(convert.FromStringArray(dir.Dir.GetPath()))
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}
		if _, err := fmt.Fprintln(x.Out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
