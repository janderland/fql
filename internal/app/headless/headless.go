package headless

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/format"
	"github.com/janderland/fdbq/parser/scanner"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type App struct {
	Flags flag.Flags
	Log   zerolog.Logger
	Out   io.Writer
}

func (x *App) Run(ctx context.Context, db facade.Transactor, queries []string) error {
	eg := engine.Engine{Tr: db, Log: x.Log}
	_, err := eg.Transact(func(eg engine.Engine) (interface{}, error) {
		for _, str := range queries {
			p := parser.New(scanner.New(strings.NewReader(str)))
			query, err := p.Parse()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse query")
			}
			if err := x.execute(ctx, eg, query); err != nil {
				return nil, errors.Wrap(err, "failed to execute query")
			}
		}
		return nil, nil
	})
	return err
}

func (x *App) execute(ctx context.Context, eg engine.Engine, query q.Query) error {
	ex := execution{ctx: ctx, app: x, eg: eg}
	query.Query(&ex)
	return ex.err
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
	str := format.Keyval(*kv)
	if _, err := fmt.Fprintln(x.Out, str); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (x *App) rangeRead(ctx context.Context, eg engine.Engine, query q.KeyValue) error {
	x.Log.Log().Interface("query", query).Msg("executing range-read query")
	for kv := range eg.RangeRead(ctx, query, x.Flags.RangeOpts()) {
		if kv.Err != nil {
			return kv.Err
		}
		str := format.Keyval(kv.KV)
		if _, err := fmt.Fprintln(x.Out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}

func (x *App) directories(ctx context.Context, eg engine.Engine, query q.Directory) error {
	x.Log.Log().Interface("query", query).Msg("executing directory query")
	for dir := range eg.Directories(ctx, query) {
		if dir.Err != nil {
			return dir.Err
		}
		str := format.Directory(convert.FromStringArray(dir.Dir.GetPath()))
		if _, err := fmt.Fprintln(x.Out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
