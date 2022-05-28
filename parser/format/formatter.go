package format

import (
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/internal"
)

type queryOp struct {
	str string
}

var _ q.QueryOperation = &queryOp{}

func (x *queryOp) ForDirectory(in q.Directory) {
	x.str = Directory(in)
}

func (x *queryOp) ForKey(in q.Key) {
	x.str = Key(in)
}

func (x *queryOp) ForKeyValue(in q.KeyValue) {
	x.str = Keyval(in)
}

type directoryOp struct {
	str string
}

var _ q.DirectoryOperation = &directoryOp{}

func (x *directoryOp) ForVariable(in q.Variable) {
	x.str = variable(in)
}

func (x *directoryOp) ForString(in q.String) {
	x.str = string(in)
}

type tupleOp struct {
	str string
}

var _ q.TupleOperation = &tupleOp{}

func (x *tupleOp) ForBigInt(q.BigInt) {
	// TODO: Implement BigInt formatting.
	panic("not implemented")
}

func (x *tupleOp) ForNil(q.Nil) {
	x.str = internal.Nil
}

func (x *tupleOp) ForMaybeMore(q.MaybeMore) {
	x.str = internal.MaybeMore
}

func (x *tupleOp) ForTuple(in q.Tuple) {
	x.str = Tuple(in)
}

func (x *tupleOp) ForInt(in q.Int) {
	x.str = integer(in)
}

func (x *tupleOp) ForUint(in q.Uint) {
	x.str = unsigned(in)
}

func (x *tupleOp) ForBool(in q.Bool) {
	x.str = boolean(in)
}

func (x *tupleOp) ForFloat(in q.Float) {
	x.str = float(in)
}

func (x *tupleOp) ForString(in q.String) {
	x.str = str(in)
}

func (x *tupleOp) ForUUID(in q.UUID) {
	x.str = uuid(in)
}

func (x *tupleOp) ForBytes(in q.Bytes) {
	x.str = hexadecimal(in)
}

func (x *tupleOp) ForVariable(in q.Variable) {
	x.str = variable(in)
}

type valueOp struct {
	str string
}

var _ q.ValueOperation = &valueOp{}

func (x *valueOp) ForNil(q.Nil) {
	x.str = internal.Nil
}

func (x *valueOp) ForClear(q.Clear) {
	x.str = internal.Clear
}

func (x *valueOp) ForTuple(in q.Tuple) {
	x.str = Tuple(in)
}

func (x *valueOp) ForInt(in q.Int) {
	x.str = integer(in)
}

func (x *valueOp) ForUint(in q.Uint) {
	x.str = unsigned(in)
}

func (x *valueOp) ForBool(in q.Bool) {
	x.str = boolean(in)
}

func (x *valueOp) ForFloat(in q.Float) {
	x.str = float(in)
}

func (x *valueOp) ForString(in q.String) {
	x.str = str(in)
}

func (x *valueOp) ForUUID(in q.UUID) {
	x.str = uuid(in)
}

func (x *valueOp) ForBytes(in q.Bytes) {
	x.str = hexadecimal(in)
}

func (x *valueOp) ForVariable(in q.Variable) {
	x.str = variable(in)
}
