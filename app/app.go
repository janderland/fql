package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
)

func Run(args []string, stdout *os.File, stderr *os.File) error {
	flags, queryStr, err := parseArgs(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return errors.Wrap(err, "failed to parse args")
	}

	query, err := parser.ParseKeyValue(queryStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse query")
	}
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get kind of query")
	}

	db, err := connectToDB()
	if err != nil {
		return errors.Wrap(err, "failed to connect to DB")
	}

	app := app{
		flags: flags,
		out:   stdout,
		eg:    engine.New(db),
	}

	if onlyDir(*query) {
		// TODO: Implement.
	}

	switch kind {
	case keyval.ConstantKind:
		if err := app.set(*query); err != nil {
			return errors.Wrap(err, "failed to execute as set query")
		}
	case keyval.ClearKind:
		if err := app.clear(*query); err != nil {
			return errors.Wrap(err, "failed to execute as clear query")
		}
	case keyval.SingleReadKind:
		if err := app.singleRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as single read query")
		}
	case keyval.RangeReadKind:
		if err := app.rangeRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as range read query")
		}
	case keyval.InvalidKind:
		return errors.New("query is invalid")
	default:
		return errors.Errorf("unexpected query kind '%v'", kind)
	}

	return nil
}

type flags struct {
	write bool
}

func parseArgs(args []string, stderr *os.File) (flags, string, error) {
	var flags flags

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.SetOutput(stderr)

	flagSet.BoolVar(&flags.write, "write", false, "allow write queries")

	if err := flagSet.Parse(args[1:]); err != nil {
		return flags, "", errors.Wrap(err, "failed to parse flags")
	}
	return flags, strings.Join(flagSet.Args(), " "), nil
}

func connectToDB() (fdb.Database, error) {
	if err := fdb.APIVersion(620); err != nil {
		return fdb.Database{}, errors.Wrap(err, "failed to specify FDB API version")
	}
	db, err := fdb.OpenDefault()
	if err != nil {
		return fdb.Database{}, errors.Wrap(err, "failed to open FDB connection")
	}
	return db, nil
}

func onlyDir(query keyval.KeyValue) bool {
	if len(query.Key.Tuple) > 0 {
		return false
	}
	if query.Value != nil {
		if tup, ok := query.Value.(keyval.Tuple); ok {
			return len(tup) == 0
		}
		return false
	}
	return true
}

type app struct {
	flags flags
	out   io.Writer
	eg    engine.Engine
}

func (a *app) set(query keyval.KeyValue) error {
	if !a.flags.write {
		return errors.New("writing isn't enabled")
	}
	return a.eg.Set(query)
}

func (a *app) clear(query keyval.KeyValue) error {
	if !a.flags.write {
		return errors.New("writing isn't enabled")
	}
	return a.eg.Clear(query)
}

func (a *app) singleRead(query keyval.KeyValue) error {
	kv, err := a.eg.SingleRead(query)
	if err != nil {
		return err
	}
	str, err := parser.FormatKeyValue(*kv)
	if err != nil {
		return errors.Wrap(err, "failed to format result")
	}
	if _, err := fmt.Fprintln(a.out, str); err != nil {
		return errors.Wrap(err, "failed to print output")
	}
	return nil
}

func (a *app) rangeRead(query keyval.KeyValue) error {
	for kv := range a.eg.RangeRead(context.Background(), query) {
		if kv.Err != nil {
			return kv.Err
		}
		str, err := parser.FormatKeyValue(kv.KV)
		if err != nil {
			return errors.Wrap(err, "failed to format result")
		}
		if _, err := fmt.Fprintln(a.out, str); err != nil {
			return errors.Wrap(err, "failed to print output")
		}
	}
	return nil
}
