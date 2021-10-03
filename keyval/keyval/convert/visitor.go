package convert

import (
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval/keyval"
	"github.com/pkg/errors"
)

type toFDB struct {
	out tuple.TupleElement
	err error
}

func (x *toFDB) VisitTuple(in q.Tuple) {
	x.out, x.err = ToFDBTuple(in)
}

func (x *toFDB) VisitNil(q.Nil) {}

func (x *toFDB) VisitInt(in q.Int) {
	x.out = int64(in)
}

func (x *toFDB) VisitUint(in q.Uint) {
	x.out = uint64(in)
}

func (x *toFDB) VisitBool(in q.Bool) {
	x.out = bool(in)
}

func (x *toFDB) VisitFloat(in q.Float) {
	x.out = float64(in)
}

func (x *toFDB) VisitBigInt(in q.BigInt) {
	x.out = big.Int(in)
}

func (x *toFDB) VisitString(in q.String) {
	x.out = string(in)
}

func (x *toFDB) VisitUUID(in q.UUID) {
	x.out = tuple.UUID(in)
}

func (x *toFDB) VisitBytes(in q.Bytes) {
	x.out = []byte(in)
}

func (x *toFDB) VisitVariable(q.Variable) {
	x.err = errors.New("cannot convert variable")
}

func (x *toFDB) VisitMaybeMore(q.MaybeMore) {
	x.err = errors.New("cannot convert maybe-more")
}
