package app

import (
	"context"
	"os"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	"github.com/janderland/fdbq/internal/app/headless"
	"github.com/janderland/fdbq/parser/format"
)

var flags *flag.Flags

func init() {
	flags = flag.SetupFlags(Fdbq)
}

var Fdbq = &cobra.Command{
	Use:   "fdbq [flags] query ...",
	Short: "fdbq is a query language for Foundation DB",
	RunE: func(_ *cobra.Command, args []string) error {
		return run(flags, args, os.Stdout, os.Stderr)
	},
}

func run(flags *flag.Flags, queries []string, stdout *os.File, stderr *os.File) error {
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

	app := headless.App{
		Format: format.New(format.Cfg{
			PrintBytes: flags.Bytes,
		}),
		Flags: *flags,
		Log:   log,
		Out:   stdout,
	}
	return app.Run(context.Background(), facade.NewTransactor(db, directory.Root()), queries)
}
