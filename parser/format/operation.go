package format

import (
	q "github.com/janderland/fdbq/keyval"
)

type dirOp struct {
	format *Format
}

var _ q.DirectoryOperation = &dirOp{}

type op struct {
	format *Format
}

// The keyval.DirectoryOperation methods are defined
// in their own struct so the ForString method could
// be handled differently. In the case of a directory,
// we want keyval.String to format without quotes.

func (x *dirOp) ForString(in q.String) {
	x.format.str.WriteString(string(in))
}

func (x *dirOp) ForVariable(in q.Variable) {
	x.format.Variable(in)
}

var (
	_ q.QueryOperation = &op{}
	_ q.TupleOperation = &op{}
	_ q.ValueOperation = &op{}
)

func (x *op) ForDirectory(in q.Directory) {
	x.format.Directory(in)
}

func (x *op) ForKey(in q.Key) {
	x.format.Key(in)
}

func (x *op) ForKeyValue(in q.KeyValue) {
	x.format.KeyValue(in)
}

func (x *op) ForVariable(in q.Variable) {
	x.format.Variable(in)
}

func (x *op) ForString(in q.String) {
	x.format.Str(in)
}

func (x *op) ForBigInt(q.BigInt) {
	// TODO: Implement BigInt formatting.
	panic("not implemented")
}

func (x *op) ForNil(in q.Nil) {
	x.format.Nil(in)
}

func (x *op) ForMaybeMore(in q.MaybeMore) {
	x.format.MaybeMore(in)
}

func (x *op) ForTuple(in q.Tuple) {
	x.format.Tuple(in)
}

func (x *op) ForInt(in q.Int) {
	x.format.Int(in)
}

func (x *op) ForUint(in q.Uint) {
	x.format.Uint(in)
}

func (x *op) ForBool(in q.Bool) {
	x.format.Bool(in)
}

func (x *op) ForFloat(in q.Float) {
	x.format.Float(in)
}

func (x *op) ForUUID(in q.UUID) {
	x.format.UUID(in)
}

func (x *op) ForBytes(in q.Bytes) {
	x.format.Bytes(in)
}

func (x *op) ForClear(in q.Clear) {
	x.format.Clear(in)
}
