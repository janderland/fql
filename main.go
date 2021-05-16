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
	query, err := parser.ParseKeyValue(strings.Join(os.Args[1:], " "))
	if err != nil {
		fail(errors.Wrap(err, "failed to parse query"))
	}
	queryStr, err := parser.FormatKeyValue(*query)
	if err != nil {
		fail(errors.Wrap(err, "failed to format query"))
	}
	fmt.Println(queryStr)

	kind, err := query.Kind()
	if err != nil {
		fail(errors.Wrap(err, "failed to get kind of query"))
	}
	if kind == keyval.InvalidKind {
		fail(errors.New("query is invalid"))
	}

	if err := fdb.APIVersion(620); err != nil {
		fail(errors.Wrap(err, "failed to specify FDB API version"))
	}
	db, err := fdb.OpenDefault()
	if err != nil {
		fail(errors.Wrap(err, "failed to open FDB connection"))
	}

	e := engine.New(db)
	switch kind {
	case keyval.ConstantKind:
		if !*write {
			fail(errors.New("writing isn't enabled"))
		}
		if err := e.Set(*query); err != nil {
			fail(errors.Wrap(err, "failed to execute as set query"))
		}

	case keyval.ClearKind:
		if !*write {
			fail(errors.New("writing isn't enabled"))
		}
		if err := e.Clear(*query); err != nil {
			fail(errors.Wrap(err, "failed to execute as clear query"))
		}

	case keyval.SingleReadKind:
		kv, err := e.SingleRead(*query)
		if err != nil {
			fail(errors.Wrap(err, "failed to execute as single read query"))
		}

		str, err := parser.FormatKeyValue(*kv)
		if err != nil {
			fail(errors.Wrap(err, "failed to format result"))
		}
		fmt.Println(str)

	case keyval.RangeReadKind:
		for kv := range e.RangeRead(context.Background(), *query) {
			if kv.Err != nil {
				fail(errors.Wrap(err, "failed to execute as range read query"))
			}
			str, err := parser.FormatKeyValue(kv.KV)
			if err != nil {
				fail(errors.Wrap(err, "failed to format result"))
			}
			fmt.Println(str)
		}
	}
}

func fail(err error) {
	fmt.Println(err)
	os.Exit(1)
}
