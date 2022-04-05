package format

import (
	q "github.com/janderland/fdbq/keyval"
)

type query struct {
	str string
}

var _ q.QueryOperation = &query{}

func (x *query) ForDirectory(dir q.Directory) {
	x.str = Directory(dir)
}

func (x *query) ForKey(key q.Key) {
	x.str = Directory(key.Directory) + Tuple(key.Tuple)
}

func (x *query) ForKeyValue(kv q.KeyValue) {
	x.str = Directory(kv.Key.Directory) + Tuple(kv.Key.Tuple)
}

type directory struct {
	str string
}

var _ q.DirectoryOperation = &directory{}

func (x *directory) ForVariable(v q.Variable) {
	x.str = Variable(v)
}

func (x *directory) ForString(str q.String) {
	x.str = string(str)
}

type tuple struct {
	str string
}

var _ q.TupleOperation = &tuple{}

func (x *tuple) ForTuple(q.Tuple) {

}

func (x *tuple) ForNil(q.Nil) {

}

func (x *tuple) ForInt(q.Int) {

}

func (x *tuple) ForUint(q.Uint) {

}

func (x *tuple) ForBool(q.Bool) {

}

func (x *tuple) ForFloat(q.Float) {

}

func (x *tuple) ForBigInt(q.BigInt) {

}

func (x *tuple) ForString(q.String) {

}

func (x *tuple) ForUUID(q.UUID) {

}

func (x *tuple) ForBytes(q.Bytes) {

}

func (x *tuple) ForVariable(q.Variable) {

}

func (x *tuple) ForMaybeMore(q.MaybeMore) {

}
