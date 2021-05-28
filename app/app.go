package app

import (
	"flag"
	"os"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser"
	"github.com/pkg/errors"
)

type flags struct {
	write bool
}

func Run(args []string, stdout *os.File, stderr *os.File) error {
	flags, queryStr, err := parseArgs(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return errors.Wrap(err, "failed to parse args")
	}

	query, onlyDir, err := parser.ParseQuery(queryStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse query")
	}

	db, err := connectToDB()
	if err != nil {
		return errors.Wrap(err, "failed to connect to DB")
	}

	app := headless{
		flags: flags,
		out:   stdout,
		eg:    engine.New(db),
	}

	if onlyDir {
		if err := app.directories(query.Key.Directory); err != nil {
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
		if err := app.set(*query); err != nil {
			return errors.Wrap(err, "failed to execute as set query")
		}
		return nil

	case keyval.ClearKind:
		if err := app.clear(*query); err != nil {
			return errors.Wrap(err, "failed to execute as clear query")
		}
		return nil

	case keyval.SingleReadKind:
		if err := app.singleRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as single read query")
		}
		return nil

	case keyval.RangeReadKind:
		if err := app.rangeRead(*query); err != nil {
			return errors.Wrap(err, "failed to execute as range read query")
		}
		return nil

	case keyval.InvalidKind:
		return errors.New("query is invalid")

	default:
		return errors.Errorf("unexpected query kind '%v'", kind)
	}
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
