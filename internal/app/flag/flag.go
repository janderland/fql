package flag

import (
	"encoding/binary"

	"github.com/janderland/fdbq/engine"
	"github.com/spf13/cobra"
)

type Flags struct {
	Cluster string
	Write   bool
	Log     bool

	Reverse bool
	Filter  bool
	Little  bool
	Bytes   bool
	Limit   int
}

func SetupFlags(cmd *cobra.Command) *Flags {
	var flags Flags

	cmd.Flags().StringVarP(&flags.Cluster, "cluster", "c", "", "path to cluster file")
	cmd.Flags().BoolVarP(&flags.Write, "write", "w", false, "allow write queries")
	cmd.Flags().BoolVar(&flags.Log, "log", false, "perform logging")

	cmd.Flags().BoolVarP(&flags.Reverse, "reverse", "r", false, "reverse range reads")
	cmd.Flags().BoolVarP(&flags.Filter, "filter", "f", false, "filter schema transgressions")
	cmd.Flags().BoolVarP(&flags.Little, "little", "l", false, "little endian value encoding")
	cmd.Flags().BoolVarP(&flags.Bytes, "bytes", "b", false, "print byte strings")
	cmd.Flags().IntVar(&flags.Limit, "limit", 0, "range read limit")

	return &flags
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
