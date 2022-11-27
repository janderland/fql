package headless

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser"
	"github.com/janderland/fdbq/parser/format"
	"github.com/janderland/fdbq/parser/scanner"
)

type App struct {
	Format format.Format
	Flags  flag.Flags
	Log    zerolog.Logger
	Out    io.Writer
}

func (x *App) Run(ctx context.Context, db facade.Transactor, queries []string) error {
	eg := engine.New(db)
	eg.Logger(x.Log)

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
	kv, err := eg.ReadSingle(query, x.Flags.SingleOpts())
	if err != nil {
		return err
	}
	if kv == nil {
		return nil
	}

	x.Format.Reset()
	x.Format.KeyValue(query)
	if _, err := fmt.Fprintln(x.Out, x.Format.String()); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (x *App) rangeRead(ctx context.Context, eg engine.Engine, query q.KeyValue) error {
	x.Log.Log().Interface("query", query).Msg("executing range-read query")
	for kv := range eg.ReadRange(ctx, query, x.Flags.RangeOpts()) {
		if kv.Err != nil {
			return kv.Err
		}

		x.Format.Reset()
		x.Format.KeyValue(kv.KV)
		if _, err := fmt.Fprintln(x.Out, x.Format.String()); err != nil {
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

		x.Format.Reset()
		x.Format.Directory(convert.FromStringArray(dir.Dir.GetPath()))
		if _, err := fmt.Fprintln(x.Out, x.Format.String()); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
