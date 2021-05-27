package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/janderland/fdbq/keyval"

	"github.com/janderland/fdbq/engine"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/pkg/errors"

	"github.com/janderland/fdbq/parser"
)

var write = flag.Bool("write", false, "allow write queries")

func main() {
	flag.Parse()

	query, err := parser.ParseKeyValue(strings.Join(flag.Args(), " "))
	if err != nil {
		fail(errors.Wrap(err, "failed to parse query"))
	}
	kind, err := query.Kind()
	if err != nil {
		fail(errors.Wrap(err, "failed to get kind of query"))
	}

	eg := engine.New(setupDB())

	if onlyDir(*query) {

	}

	switch kind {
	case keyval.ConstantKind:
		if err := set(eg, *query); err != nil {
			fail(errors.Wrap(err, "failed to execute as set query"))
		}
	case keyval.ClearKind:
		if err := clear(eg, *query); err != nil {
			fail(errors.Wrap(err, "failed to execute as clear query"))
		}
	case keyval.SingleReadKind:
		if err := singleRead(eg, *query); err != nil {
			fail(errors.Wrap(err, "failed to execute as single read query"))
		}
	case keyval.RangeReadKind:
		if err := rangeRead(eg, *query); err != nil {
			fail(errors.Wrap(err, "failed to execute as range read query"))
		}
	case keyval.InvalidKind:
		fail(errors.New("query is invalid"))
	default:
		panic(errors.Errorf("unexpected query kind '%v'", kind))
	}
}

func setupDB() fdb.Database {
	if err := fdb.APIVersion(620); err != nil {
		fail(errors.Wrap(err, "failed to specify FDB API version"))
	}
	db, err := fdb.OpenDefault()
	if err != nil {
		fail(errors.Wrap(err, "failed to open FDB connection"))
	}
	return db
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

func set(e engine.Engine, query keyval.KeyValue) error {
	if !*write {
		return errors.New("writing isn't enabled")
	}
	return e.Set(query)
}

func clear(e engine.Engine, query keyval.KeyValue) error {
	if !*write {
		return errors.New("writing isn't enabled")
	}
	return e.Clear(query)
}

func singleRead(e engine.Engine, query keyval.KeyValue) error {
	kv, err := e.SingleRead(query)
	if err != nil {
		return err
	}
	str, err := parser.FormatKeyValue(*kv)
	if err != nil {
		return errors.Wrap(err, "failed to format result")
	}
	fmt.Println(str)
	return nil
}

func rangeRead(e engine.Engine, query keyval.KeyValue) error {
	for kv := range e.RangeRead(context.Background(), query) {
		if kv.Err != nil {
			return kv.Err
		}
		str, err := parser.FormatKeyValue(kv.KV)
		if err != nil {
			return errors.Wrap(err, "failed to format result")
		}
		fmt.Println(str)
	}
	return nil
}

func fail(err error) {
	fmt.Println(err)
	os.Exit(1)
}
