package iterator

import (
	"math"
	"math/big"
	"testing"

	q "github.com/janderland/fdbq/keyval/keyval"
	"github.com/stretchr/testify/assert"
)

func TestReadTuple(t *testing.T) {
	in := q.Tuple{
		q.Nil{},
		q.Bool(true),
		q.String("hello world"),
		q.Int(math.MaxInt64),
		q.Uint(math.MaxUint64),
		q.BigInt(*big.NewInt(math.MaxInt64)),
		q.Float(math.MaxFloat64),
		q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9},
		q.Bytes{0xFF, 0xAA, 0x00},
		q.Tuple{q.Bool(true), q.Int(10)},
	}

	var out q.Tuple
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
	in := q.Tuple{q.Bool(true), q.Bool(false)}
	var out []q.Bool
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustBool())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []q.Bool{true, false}, out)
}

func TestTupleIterator_String(t *testing.T) {
	in := q.Tuple{q.String("hello"), q.String("goodbye"), q.String("world")}
	var out []q.String
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustString())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []q.String{"hello", "goodbye", "world"}, out)
}

func TestTupleIterator_Int(t *testing.T) {
	in := q.Tuple{q.Int(23), q.Int(-32)}
	var out []q.Int
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustInt())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []q.Int{23, -32}, out)
}

func TestTupleIterator_Uint(t *testing.T) {
	in := q.Tuple{q.Uint(23), q.Uint(32)}
	var out []q.Uint
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustUint())
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []q.Uint{23, 32}, out)
}

func TestTupleIterator_BigInt(t *testing.T) {
	// This value is needed because we can't overflow
	// a negative constant into a uint64.
	neg := int64(-32)

	in := q.Tuple{q.Uint(23), q.Uint(neg), q.Int(23), q.Int(-32), q.BigInt(*big.NewInt(10))}
	var out []q.BigInt
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, iter.MustBigInt())
		}
		return nil
	})
	assert.NoError(t, err)

	bigBoi := big.NewInt(0)
	bigBoi.SetUint64(uint64(neg))
	assert.Equal(t, []q.BigInt{q.BigInt(*big.NewInt(23)), q.BigInt(*bigBoi), q.BigInt(*big.NewInt(23)), q.BigInt(*big.NewInt(-32)), q.BigInt(*big.NewInt(10))}, out)
}

func TestTupleIterator_Float(t *testing.T) {
	in := q.Tuple{q.Float(12.3), q.Float(-55.234)}
	var out []float64
	err := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		for range in {
			out = append(out, float64(iter.MustFloat()))
		}
		return nil
	})
	assert.NoError(t, err)
	assert.InEpsilonSlice(t, []float64{12.3, -55.234}, out, 0.0001)
}
