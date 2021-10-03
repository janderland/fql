package old

import (
	"math"
	"math/big"
	"testing"

	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/stretchr/testify/assert"
)

func TestReadTuple(t *testing.T) {
	in := Tuple{
		nil,
		true,
		"hello world",
		int64(math.MaxInt64),
		uint64(math.MaxUint64),
		big.NewInt(math.MaxInt64),
		math.MaxFloat64,
		tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9},
		[]byte{0xFF, 0xAA, 0x00},
		Tuple{true, int64(10)},
	}

	var out Tuple
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		out = append(out, iter.Any())
		out = append(out, iter.MustBool())
		out = append(out, iter.MustString())
		out = append(out, iter.MustInt())
		out = append(out, iter.MustUint())
		out = append(out, iter.MustBigInt())
		out = append(out, iter.MustFloat())
		out = append(out, iter.MustUUID())
		out = append(out, iter.MustBytes())
		out = append(out, iter.MustTuple())
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestTupleIterator_Bool(t *testing.T) {
	in := Tuple{true, false}
	var out []bool
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustBool())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []bool{true, false}, out)
}

func TestTupleIterator_String(t *testing.T) {
	in := Tuple{"hello", "goodbye", "world"}
	var out []string
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustString())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "goodbye", "world"}, out)
}

func TestTupleIterator_Int(t *testing.T) {
	in := Tuple{23, int64(-32)}
	var out []int64
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustInt())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int64{23, -32}, out)
}

func TestTupleIterator_Uint(t *testing.T) {
	// This value is needed because we can't overflow
	// a negative constant into a uint64 at the final
	// assert, but we can overflow a value.
	neg := int64(-32)

	in := Tuple{uint(23), uint64(32), 23, neg}
	var out []uint64
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustUint())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []uint64{23, 32, 23, uint64(neg)}, out)
}

func TestTupleIterator_BigInt(t *testing.T) {
	// This value is needed because we can't overflow
	// a negative constant into a uint64.
	neg := int64(-32)

	in := Tuple{uint(23), uint64(neg), 23, -32, big.NewInt(10)}
	var out []*big.Int
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustBigInt())
		}
		return nil
	})
	assert.NoError(t, err)

	bigBoi := big.NewInt(0)
	bigBoi.SetUint64(uint64(neg))
	assert.Equal(t, []*big.Int{big.NewInt(23), bigBoi, big.NewInt(23), big.NewInt(-32), big.NewInt(10)}, out)
}

func TestTupleIterator_Float(t *testing.T) {
	in := Tuple{float32(12.3), float64(-55.234)}
	var out []float64
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustFloat())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.InEpsilonSlice(t, []float64{12.3, -55.234}, out, 0.0001)
}
