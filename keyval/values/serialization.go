package values

import (
	"encoding/binary"
	"math"

	q "github.com/janderland/fql/keyval"
	"github.com/janderland/fql/keyval/convert"
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
	if x.vstamp {
		x.packed, x.err = tup.PackWithVersionstamp(nil)
		return
	} 
	x.packed = tup.Pack()
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

func (x *serialization) ForVStamp(e q.VStamp) {
	// TX version   - 10B
	// User version - 2B
	x.packed = make([]byte, 12)
	copy(x.packed[0:10], e.TxVersion[:])
	binary.LittleEndian.PutUint16(x.packed[10:12], e.UserVersion)
}

func (x *serialization) ForVStampFuture(e q.VStampFuture) {
	// TX version   - 10B
	// User version - 2B
	// TX position  - 4B
	x.packed = make([]byte, 16)
	binary.LittleEndian.PutUint16(x.packed[10:12], e.UserVersion)
}

func (x *serialization) ForNil(q.Nil) {}

func (x *serialization) ForVariable(q.Variable) {
	x.err = errors.New("cannot serialize a variable")
}

func (x *serialization) ForClear(q.Clear) {
	x.err = errors.New("cannot serialize a clear")
}
