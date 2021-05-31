package app

import (
	"context"
	"fmt"
	"os"

	"github.com/janderland/fdbq/parser"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/app/flag"
	"github.com/janderland/fdbq/app/headless"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func Run() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout *os.File, stderr *os.File) error {
	flags, queries, err := flag.Parse(args, stderr)
	if err != nil {
		return errors.Wrap(err, "failed to parse args")
	}
	if flags == nil {
		return nil
	}

	log := zerolog.Nop()
	if flags.Log {
		writer := zerolog.ConsoleWriter{Out: stderr}
		writer.FormatLevel = func(_ interface{}) string { return "" }
		log = zerolog.New(writer).With().Timestamp().Logger()
	}

	log.Log().Str("cluster file", flags.Cluster).Msg("connecting to DB")
	if err := fdb.APIVersion(620); err != nil {
		return errors.Wrap(err, "failed to set FDB API version")
	}
	db, err := fdb.OpenDatabase(flags.Cluster)
	if err != nil {
		return errors.Wrap(err, "failed to connect to DB")
	}

	app := headless.New(log.WithContext(context.Background()), *flags, stdout, db)
	for _, query := range queries {
		if err := app.Query(query); err != nil {
			switch err := err.(type) {
			case parser.Error:
				_, _ = fmt.Fprint(stderr, err.SPrint())
				return errors.New("")
			}
			return errors.Wrapf(err, "failed to execute '%s'", query)
		}
	}
	return nil
}
