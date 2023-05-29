package flag

import (
	"encoding/binary"

	"github.com/spf13/cobra"

	"github.com/janderland/fdbq/engine"
)

type Flags struct {
	Cluster string
	Write   bool
	Log     bool

	Queries []string
	Reverse bool
	Strict  bool
	Little  bool
	Bytes   bool
	Limit   int
}

func SetupFlags(cmd *cobra.Command) *Flags {
	var flags Flags

	cmd.Flags().StringVarP(&flags.Cluster, "cluster", "c", "", "path to cluster file")
	cmd.Flags().BoolVarP(&flags.Write, "write", "w", false, "allow write queries")
	cmd.Flags().BoolVar(&flags.Log, "log", false, "perform debug logging")

	cmd.Flags().StringArrayVarP(&flags.Queries, "query", "q", nil, "execute query non-interactively")
	cmd.Flags().BoolVarP(&flags.Reverse, "reverse", "r", false, "query range-reads in reverse order")
	cmd.Flags().BoolVarP(&flags.Strict, "strict", "s", false, "throw an error if a KV is read which doesn't match the schema")
	cmd.Flags().BoolVarP(&flags.Little, "little", "l", false, "encode/decode values as little endian instead of big endian")
	cmd.Flags().BoolVarP(&flags.Bytes, "bytes", "b", false, "print full byte strings instead of just their length")
	cmd.Flags().IntVar(&flags.Limit, "limit", 0, "limit the number of KVs read in range-reads")

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
		Filter: !x.Strict,
	}
}

func (x *Flags) RangeOpts() engine.RangeOpts {
	return engine.RangeOpts{
		Reverse: x.Reverse,
		Filter:  !x.Strict,
		Limit:   x.Limit,
	}
}
