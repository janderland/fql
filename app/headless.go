package app

import (
	"context"
	"fmt"
	"io"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
)

type headless struct {
	flags flags
	out   io.Writer
	eg    engine.Engine
}

func (h *headless) set(query keyval.KeyValue) error {
	if !h.flags.write {
		return errors.New("writing isn't enabled")
	}
	return h.eg.Set(query)
}

func (h *headless) clear(query keyval.KeyValue) error {
	if !h.flags.write {
		return errors.New("writing isn't enabled")
	}
	return h.eg.Clear(query)
}

func (h *headless) singleRead(query keyval.KeyValue) error {
	kv, err := h.eg.SingleRead(query)
	if err != nil {
		return err
	}
	str, err := parser.FormatKeyValue(*kv)
	if err != nil {
		return errors.Wrap(err, "failed to format result")
	}
	if _, err := fmt.Fprintln(h.out, str); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (h *headless) rangeRead(query keyval.KeyValue) error {
	for kv := range h.eg.RangeRead(context.Background(), query) {
		if kv.Err != nil {
			return kv.Err
		}
		str, err := parser.FormatKeyValue(kv.KV)
		if err != nil {
			return errors.Wrap(err, "failed to format result")
		}
		if _, err := fmt.Fprintln(h.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}

func (h *headless) directories(query keyval.Directory) error {
	for msg := range h.eg.Directories(context.Background(), query) {
		if msg.Err != nil {
			return msg.Err
		}
		str, err := parser.FormatDirectory(keyval.FromStringArray(msg.Dir.GetPath()))
		if err != nil {
			return errors.Wrap(err, "failed to format result")
		}
		if _, err := fmt.Fprintln(h.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
