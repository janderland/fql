package flag

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"github.com/janderland/fdbq/engine"
	"github.com/pkg/errors"
)

type Flags struct {
	Cluster string
	Write   bool
	Log     bool

	Reverse bool
	Filter  bool
	Little  bool
	Limit   int
}

func setupFlagSet(output *os.File) (*Flags, *flag.FlagSet) {
	var flags Flags

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {
		_, _ = fmt.Fprint(fs.Output(), "fdbq [flags] query1 [query2...]\n\n")
		fs.PrintDefaults()
	}

	fs.SetOutput(output)

	fs.StringVar(&flags.Cluster, "cluster", "", "path to cluster file")
	fs.BoolVar(&flags.Write, "write", false, "allow write queries")
	fs.BoolVar(&flags.Log, "log", false, "perform logging")

	fs.BoolVar(&flags.Reverse, "reverse", false, "reverse range reads")
	fs.BoolVar(&flags.Filter, "filter", false, "filter schema transgressions")
	fs.BoolVar(&flags.Little, "little", false, "little endian value encoding")
	fs.IntVar(&flags.Limit, "limit", 0, "range read limit")

	return &flags, fs
}

func Parse(args []string, stderr *os.File) (*Flags, []string, error) {
	flags, flagSet := setupFlagSet(stderr)
	if err := flagSet.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to parse flags")
	}
	return flags, flagSet.Args(), nil
}

func (x *Flags) ByteOrder() binary.ByteOrder {
	if x.Little {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (x *Flags) SingleOpts() engine.SingleOpts {
	return engine.SingleOpts{
		ByteOrder: x.ByteOrder(),
		Filter:    x.Filter,
	}
}

func (x *Flags) RangeOpts() engine.RangeOpts {
	return engine.RangeOpts{
		ByteOrder: x.ByteOrder(),
		Reverse:   x.Reverse,
		Filter:    x.Filter,
		Limit:     x.Limit,
	}
}
