package headless

import (
	"context"
	"fmt"
	"io"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/internal/app/flag"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Headless struct {
	flags flag.Flags
	log   *zerolog.Logger
	out   io.Writer
	eg    engine.Engine
}

func New(ctx context.Context, flags flag.Flags, out io.Writer, db fdb.Transactor) Headless {
	return Headless{
		flags: flags,
		log:   zerolog.Ctx(ctx),
		out:   out,
		eg:    engine.New(ctx, db),
	}
}

func (h *Headless) Query(str string) error {
	query, onlyDir, err := parser.ParseQuery(str)
	if err != nil {
		return errors.Wrap(err, "failed to parse query")
	}

	if onlyDir {
		if err := h.directories(query.Key.Directory); err != nil {
			return errors.Wrap(err, "failed to execute as directory query")
		}
		return nil
	}

	switch c := class.Classify(*query); c {
	case class.Constant:
		if err := h.set(*query); err != nil {
			return errors.Wrap(err, "failed to execute as set query")
		}
		return nil

	case class.Clear:
		if err := h.clear(*query); err != nil {
			return errors.Wrap(err, "failed to execute as clear query")
		}
		return nil

	case class.SingleRead:
		if err := h.singleRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as single read query")
		}
		return nil

	case class.RangeRead:
		if err := h.rangeRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as range read query")
		}
		return nil

	default:
		return errors.Errorf("unexpected query class '%v'", c)
	}
}

func (h *Headless) set(query q.KeyValue) error {
	if !h.flags.Write {
		return errors.New("writing isn't enabled")
	}
	h.log.Log().Interface("query", query).Msg("executing set query")
	return h.eg.Set(query, h.flags.ByteOrder())
}

func (h *Headless) clear(query q.KeyValue) error {
	if !h.flags.Write {
		return errors.New("writing isn't enabled")
	}
	h.log.Log().Interface("query", query).Msg("executing clear query")
	return h.eg.Clear(query)
}

func (h *Headless) singleRead(query q.KeyValue) error {
	h.log.Log().Interface("query", query).Msg("executing single-read query")
	kv, err := h.eg.SingleRead(query, h.flags.ByteOrder())
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
	if _, err := fmt.Fprintln(h.out, str); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (h *Headless) rangeRead(query q.KeyValue) error {
	h.log.Log().Interface("query", query).Msg("executing range-read query")
	for kv := range h.eg.RangeRead(context.Background(), query, h.flags.RangeOpts()) {
		if kv.Err != nil {
			return kv.Err
		}
		str, err := parser.FormatKeyValue(kv.KV)
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}
		if _, err := fmt.Fprintln(h.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}

func (h *Headless) directories(query q.Directory) error {
	h.log.Log().Interface("query", query).Msg("executing directory query")
	for msg := range h.eg.Directories(context.Background(), query) {
		if msg.Err != nil {
			return msg.Err
		}
		str, err := parser.FormatDirectory(convert.FromStringArray(msg.Dir.GetPath()))
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}
		if _, err := fmt.Fprintln(h.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
