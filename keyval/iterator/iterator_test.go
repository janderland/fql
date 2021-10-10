package iterator

import (
	"math"
	"math/big"
	"testing"

	q "github.com/janderland/fdbq/keyval"
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

func TestConversionError(t *testing.T) {
	in := q.Tuple{q.Bool(true)}
	out := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		_ = iter.MustString()
		return nil
	})
	assert.IsType(t, ConversionError{}, out)
}

func TestShortTupleError(t *testing.T) {
	in := q.Tuple{q.Int(10)}
	out := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
		_ = iter.MustInt()
		_ = iter.MustInt()
		return nil
	})
	assert.Equal(t, ShortTupleError, out)
}

func TestLongTupleError(t *testing.T) {
	in := q.Tuple{q.Bytes{0xFF, 0xAC}, q.String("hi")}

	t.Run("error", func(t *testing.T) {
		out := ReadTuple(in, AllErrors, func(iter *TupleIterator) error {
			_ = iter.MustBytes()
			return nil
		})
		assert.Equal(t, LongTupleError, out)
	})

	t.Run("allow", func(t *testing.T) {
		out := ReadTuple(in, AllowLong, func(iter *TupleIterator) error {
			_ = iter.MustBytes()
			return nil
		})
		assert.NoError(t, out)
	})
}
