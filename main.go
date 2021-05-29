package main

import (
	"fmt"
	"os"

	"github.com/apple/foundationdb/bindings/go/src/fdb"

	"github.com/pkg/errors"

	"github.com/janderland/fdbq/app"
)

func main() {
	fdb.MustAPIVersion(620)

	if err := app.Run(os.Args, os.Stdout, os.Stderr); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "%v\n", err); err != nil {
			panic(errors.Wrap(err, "failed to display error"))
		}
		os.Exit(1)
	}
}
