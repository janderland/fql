package format

import (
	"strconv"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/parser/internal"
)

type queryOp struct {
	str string
}

var _ q.QueryOperation = &queryOp{}

func (x *queryOp) ForDirectory(in q.Directory) {
	x.str = directory(in)
}

func (x *queryOp) ForKey(in q.Key) {
	x.str = directory(in.Directory) + tuple(in.Tuple)
}

func (x *queryOp) ForKeyValue(in q.KeyValue) {
	x.str = directory(in.Key.Directory) + tuple(in.Key.Tuple)
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

func (x *tupleOp) ForTuple(tup q.Tuple) {
	x.str = tuple(tup)
}

func (x *tupleOp) ForNil(q.Nil) {
	x.str = internal.Nil
}

func (x *tupleOp) ForInt(in q.Int) {
	x.str = strconv.FormatInt(int64(in), 10)
}

func (x *tupleOp) ForUint(in q.Uint) {
	x.str = strconv.FormatUint(uint64(in), 10)
}

func (x *tupleOp) ForBool(in q.Bool) {
	if in {
		x.str = internal.True
	} else {
		x.str = internal.False
	}
}

func (x *tupleOp) ForFloat(in q.Float) {
	x.str = strconv.FormatFloat(float64(in), 'g', 10, 64)
}

func (x *tupleOp) ForBigInt(q.BigInt) {
	// TODO: Implement BigInt formatting.
	panic("not implemented")
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

func (x *tupleOp) ForMaybeMore(q.MaybeMore) {
	x.str = internal.MaybeMore
}
