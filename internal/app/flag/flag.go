package flag

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"github.com/janderland/fdbq/engine/stream"

	"github.com/pkg/errors"
)

type Flags struct {
	Cluster string
	Write   bool
	Log     bool

	Reverse bool
	Little  bool
	Limit   int
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
	flagSet.BoolVar(&flags.Write, "write", false, "allow write queries")
	flagSet.BoolVar(&flags.Log, "log", false, "perform logging")

	flagSet.BoolVar(&flags.Reverse, "reverse", false, "reverse range reads")
	flagSet.BoolVar(&flags.Little, "little", false, "little endian value encoding")
	flagSet.IntVar(&flags.Limit, "limit", 0, "range read limit")

	if err := flagSet.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to parse flags")
	}
	return &flags, flagSet.Args(), nil
}

func (x *Flags) ByteOrder() binary.ByteOrder {
	if x.Little {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (x *Flags) RangeOpts() stream.RangeOpts {
	return stream.RangeOpts{
		ByteOrder: x.ByteOrder(),
		Reverse:   x.Reverse,
		Limit:     x.Limit,
	}
}
