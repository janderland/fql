package convert

import (
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type conversion struct {
	out tuple.TupleElement
	err error
}

func (x *conversion) VisitTuple(in q.Tuple) {
	x.out, x.err = ToFDBTuple(in)
}

func (x *conversion) VisitNil(q.Nil) {}

func (x *conversion) VisitInt(in q.Int) {
	x.out = int64(in)
}

func (x *conversion) VisitUint(in q.Uint) {
	x.out = uint64(in)
}

func (x *conversion) VisitBool(in q.Bool) {
	x.out = bool(in)
}

func (x *conversion) VisitFloat(in q.Float) {
	x.out = float64(in)
}

func (x *conversion) VisitBigInt(in q.BigInt) {
	x.out = big.Int(in)
}

func (x *conversion) VisitString(in q.String) {
	x.out = string(in)
}

func (x *conversion) VisitUUID(in q.UUID) {
	x.out = tuple.UUID(in)
}

func (x *conversion) VisitBytes(in q.Bytes) {
	x.out = []byte(in)
}

func (x *conversion) VisitVariable(q.Variable) {
	x.err = errors.New("cannot convert variable")
}

func (x *conversion) VisitMaybeMore(q.MaybeMore) {
	x.err = errors.New("cannot convert maybe-more")
}
