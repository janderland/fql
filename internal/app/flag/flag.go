package flag

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type Flags struct {
	Cluster string
	Little  bool
	Write   bool
	Log     bool
}

func Parse(args []string, stderr *os.File) (*Flags, []string, error) {
	var flags Flags

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	flagSet.Usage = func() {
		_, _ = fmt.Fprint(flagSet.Output(), "fdbq [flags] query1 [query2...]\n\n")
		flagSet.PrintDefaults()
	}

	flagSet.StringVar(&flags.Cluster, "cluster", "", "path to cluster file")
	flagSet.BoolVar(&flags.Little, "little", false, "little endian value encoding")
	flagSet.BoolVar(&flags.Write, "write", false, "allow write queries")
	flagSet.BoolVar(&flags.Log, "log", false, "perform logging")

	if err := flagSet.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to parse flags")
	}
	return &flags, flagSet.Args(), nil
}
