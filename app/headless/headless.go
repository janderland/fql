package headless

import (
	"context"
	"fmt"
	"io"

	"github.com/janderland/fdbq/app/flag"

	"github.com/apple/foundationdb/bindings/go/src/fdb"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
)

type Headless struct {
	flags flag.Flags
	out   io.Writer
	eg    engine.Engine
}

func New(flags flag.Flags, out io.Writer, db fdb.Transactor) Headless {
	return Headless{flags: flags, out: out, eg: engine.New(db)}
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

	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get kind of query")
	}

	switch kind {
	case keyval.ConstantKind:
		if err := h.set(*query); err != nil {
			return errors.Wrap(err, "failed to execute as set query")
		}
		return nil

	case keyval.ClearKind:
		if err := h.clear(*query); err != nil {
			return errors.Wrap(err, "failed to execute as clear query")
		}
		return nil

	case keyval.SingleReadKind:
		if err := h.singleRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as single read query")
		}
		return nil

	case keyval.RangeReadKind:
		if err := h.rangeRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as range read query")
		}
		return nil

	case keyval.InvalidKind:
		return errors.New("query is invalid")

	default:
		return errors.Errorf("unexpected query kind '%v'", kind)
	}
}

func (h *Headless) set(query keyval.KeyValue) error {
	if !h.flags.Write {
		return errors.New("writing isn't enabled")
	}
	return h.eg.Set(query)
}

func (h *Headless) clear(query keyval.KeyValue) error {
	if !h.flags.Write {
		return errors.New("writing isn't enabled")
	}
	return h.eg.Clear(query)
}

func (h *Headless) singleRead(query keyval.KeyValue) error {
	kv, err := h.eg.SingleRead(query)
	if err != nil {
		return err
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

func (h *Headless) rangeRead(query keyval.KeyValue) error {
	for kv := range h.eg.RangeRead(context.Background(), query) {
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

func (h *Headless) directories(query keyval.Directory) error {
	for msg := range h.eg.Directories(context.Background(), query) {
		if msg.Err != nil {
			return msg.Err
		}
		str, err := parser.FormatDirectory(keyval.FromStringArray(msg.Dir.GetPath()))
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}
		if _, err := fmt.Fprintln(h.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
