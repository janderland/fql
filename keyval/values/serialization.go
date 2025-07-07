package values

import (
	"encoding/binary"
	"math"

	q "github.com/janderland/fql/keyval"
	"github.com/janderland/fql/keyval/convert"
	"github.com/janderland/fql/keyval/tuple"
	"github.com/pkg/errors"
)

type serialization struct {
	order  binary.ByteOrder
	packed []byte
	vstamp bool
	err    error
}

var _ q.ValueOperation = &serialization{}

func (x *serialization) ForTuple(v q.Tuple) {
	tup, err := convert.ToFDBTuple(v)
	if err != nil {
		x.err = errors.Wrap(err, "failed to convert to FDB tuple")
		return
	}
	if x.vstamp = tuple.HasVStampFuture(v); x.vstamp {
		x.packed, x.err = tup.PackWithVersionstamp(nil)
		return
	} else {
		x.packed = tup.Pack()
  }
}

func (x *serialization) ForInt(v q.Int) {
	x.packed = make([]byte, 8)
	x.order.PutUint64(x.packed, uint64(v))
}

func (x *serialization) ForUint(v q.Uint) {
	x.packed = make([]byte, 8)
	x.order.PutUint64(x.packed, uint64(v))
}

func (x *serialization) ForBool(v q.Bool) {
	if v {
		x.packed = []byte{1}
	} else {
		x.packed = []byte{0}
	}
}

func (x *serialization) ForFloat(v q.Float) {
	x.packed = make([]byte, 8)
	x.order.PutUint64(x.packed, math.Float64bits(float64(v)))
}

func (x *serialization) ForString(v q.String) {
	x.packed = []byte(v)
}

func (x *serialization) ForUUID(v q.UUID) {
	x.packed = v[:]
}

func (x *serialization) ForBytes(v q.Bytes) {
	x.packed = v
}

func (x *serialization) ForNil(_ q.Nil) {}

func (x *serialization) ForVariable(_ q.Variable) {
	x.err = errors.New("cannot serialize a variable")
}

func (x *serialization) ForClear(_ q.Clear) {
	x.err = errors.New("cannot serialize a clear")
}
