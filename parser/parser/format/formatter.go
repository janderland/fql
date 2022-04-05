package format

import (
	"strconv"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/parser/internal"
)

type query struct {
	str string
}

var _ q.QueryOperation = &query{}

func (x *query) ForDirectory(in q.Directory) {
	x.str = Directory(in)
}

func (x *query) ForKey(in q.Key) {
	x.str = Directory(in.Directory) + Tuple(in.Tuple)
}

func (x *query) ForKeyValue(in q.KeyValue) {
	x.str = Directory(in.Key.Directory) + Tuple(in.Key.Tuple)
}

type directory struct {
	str string
}

var _ q.DirectoryOperation = &directory{}

func (x *directory) ForVariable(in q.Variable) {
	x.str = Variable(in)
}

func (x *directory) ForString(in q.String) {
	x.str = string(in)
}

type tuple struct {
	str string
}

var _ q.TupleOperation = &tuple{}

func (x *tuple) ForTuple(tup q.Tuple) {
	x.str = Tuple(tup)
}

func (x *tuple) ForNil(q.Nil) {
	x.str = internal.Nil
}

func (x *tuple) ForInt(in q.Int) {
	x.str = strconv.FormatInt(int64(in), 10)
}

func (x *tuple) ForUint(in q.Uint) {
	x.str = strconv.FormatUint(uint64(in), 10)
}

func (x *tuple) ForBool(in q.Bool) {
	if in {
		x.str = internal.True
	} else {
		x.str = internal.False
	}
}

func (x *tuple) ForFloat(in q.Float) {
	x.str = strconv.FormatFloat(float64(in), 'g', 10, 64)
}

func (x *tuple) ForBigInt(q.BigInt) {
	// TODO: Implement BigInt formatting.
	panic("not implemented")
}

func (x *tuple) ForString(in q.String) {
	x.str = String(in)
}

func (x *tuple) ForUUID(in q.UUID) {
	x.str = UUID(in)
}

func (x *tuple) ForBytes(in q.Bytes) {
	x.str = Hex(in)
}

func (x *tuple) ForVariable(in q.Variable) {
	x.str = Variable(in)
}

func (x *tuple) ForMaybeMore(q.MaybeMore) {
	x.str = internal.MaybeMore
}
