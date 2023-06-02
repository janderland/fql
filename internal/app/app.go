package app

import (
	"context"
	"os"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	"github.com/janderland/fdbq/internal/app/fullscreen"
	"github.com/janderland/fdbq/internal/app/headless"
	"github.com/janderland/fdbq/parser/format"
)

var (
	// Version is meant to be set via build flags
	// and defines the version printed for the
	// `-v` flag.
	Version string

	flags *flag.Flags
)

func init() {
	flags = flag.SetupFlags(Fdbq)
}

var Fdbq = &cobra.Command{
	Use:     "fdbq [flags] query ...",
	Short:   "fdbq is a query language for Foundation DB",
	Version: Version,

	RunE: func(_ *cobra.Command, _ []string) error {
		log := zerolog.Nop()
		if flags.Log {
			writer := zerolog.ConsoleWriter{Out: os.Stderr}
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

		eg := engine.New(
			facade.NewTransactor(db, directory.Root()),
			engine.ByteOrder(flags.ByteOrder()),
			engine.Logger(log),
		)

		kvFmt := format.New(format.Cfg{PrintBytes: flags.Bytes})

		if len(flags.Queries) != 0 {
			app := headless.App{
				Engine: eg,
				Format: kvFmt,
				Flags:  *flags,
				Out:    os.Stdout,
			}
			return app.Run(context.Background(), flags.Queries)
		}
		app := fullscreen.App{
			Engine: eg,
			Format: kvFmt,
			Flags:  *flags,
			Log:    log,
			Out:    os.Stdout,
		}
		return app.Run(context.Background())
	},
}
