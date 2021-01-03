package keyval

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func TestParseTuple(t *testing.T) {
	in := tup.Tuple{
		nil,
		true,
		"hello world",
		int64(math.MaxInt64),
		uint64(math.MaxUint64),
		big.NewInt(math.MaxInt64),
		math.MaxFloat64,
		tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9},
		[]byte{0xFF, 0xAA, 0x00},
	}

	out := make(tup.Tuple, len(in))
	err := ParseTuple(in, func(p *TupleParser) {
		out[0] = p.Any()
		out[1] = p.Bool()
		out[2] = p.String()
		out[3] = p.Int()
		out[4] = p.Uint()
		out[5] = p.BigInt()
		out[6] = p.Float()
		out[7] = p.UUID()
		out[8] = p.Bytes()
	})

	assert.NoError(t, err)
	assert.Equal(t, in, out)
}
