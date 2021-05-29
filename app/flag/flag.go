package flag

import (
	"flag"
	"os"

	"github.com/pkg/errors"
)

type Flags struct {
	Write   bool
	Cluster string
}

func Parse(args []string, stderr *os.File) (*Flags, []string, error) {
	var flags Flags

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.SetOutput(stderr)

	flagSet.BoolVar(&flags.Write, "write", false, "allow write queries")
	flagSet.StringVar(&flags.Cluster, "cluster", "", "path to cluster file")

	if err := flagSet.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// TODO: Reduce possible return values.
			// Returning nil products with nil errors to specify
			// that the help message was requested may be confusing.
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to parse flags")
	}
	return &flags, flagSet.Args(), nil
}
