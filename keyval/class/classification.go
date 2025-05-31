package class

import q "github.com/janderland/fql/keyval"

var (
	_ q.DirectoryOperation = &dirClassification{}
	_ q.TupleOperation     = &tupClassification{}
	_ q.ValueOperation     = &valClassification{}
)

type dirClassification struct {
	hasVariable bool
}

func (x *dirClassification) orFields(c character) character {
	c.hasVariable = c.hasVariable || x.hasVariable
	return c
}

func (x *dirClassification) ForString(q.String) {}

func (x *dirClassification) ForVariable(q.Variable) {
	x.hasVariable = true
}

type tupClassification struct {
	vstampFutures int
	hasVariable   bool
	hasNil        bool
}

func (x *tupClassification) orFields(c character) character {
	c.vstampFutures = c.vstampFutures + x.vstampFutures
	c.hasVariable = c.hasVariable || x.hasVariable
	c.hasNil = c.hasNil || x.hasNil
	return c
}

func (x *tupClassification) ForTuple(e q.Tuple) {
	class := classifyTuple(e)
	x.vstampFutures = x.vstampFutures + class.vstampFutures
	x.hasVariable = x.hasVariable || class.hasVariable
	x.hasNil = x.hasNil || class.hasNil
}

func (x *tupClassification) ForVariable(q.Variable) {
	x.hasVariable = true
}

func (x *tupClassification) ForMaybeMore(q.MaybeMore) {
	x.hasVariable = true
}

func (x *tupClassification) ForNil(q.Nil) {}

func (x *tupClassification) ForInt(q.Int) {}

func (x *tupClassification) ForUint(q.Uint) {}

func (x *tupClassification) ForBool(q.Bool) {}

func (x *tupClassification) ForFloat(q.Float) {}

// TODO: Add support for BigInt.
/*
func (x *tupClassification) ForBigInt(q.BigInt) {}
*/

func (x *tupClassification) ForString(q.String) {}

func (x *tupClassification) ForUUID(q.UUID) {}

func (x *tupClassification) ForBytes(q.Bytes) {}

func (x *tupClassification) ForVStamp(q.VStamp) {}

func (x *tupClassification) ForVStampFuture(q.VStampFuture) {
	x.vstampFutures++
}

type valClassification struct {
	vstampFutures int
	hasVariable   bool
	hasClear      bool
}

func (x *valClassification) orFields(c character) character {
	c.vstampFutures = c.vstampFutures + x.vstampFutures
	c.hasVariable = c.hasVariable || x.hasVariable
	c.hasClear = c.hasClear || x.hasClear
	return c
}

func (x *valClassification) ForTuple(e q.Tuple) {
	class := classifyTuple(e)
	x.vstampFutures = x.vstampFutures + class.vstampFutures
	x.hasVariable = x.hasVariable || class.hasVariable
}

func (x *valClassification) ForVariable(q.Variable) {
	x.hasVariable = true
}

func (x *valClassification) ForClear(q.Clear) {
	x.hasClear = true
}

func (x *valClassification) ForNil(q.Nil) {}

func (x *valClassification) ForInt(q.Int) {}

func (x *valClassification) ForUint(q.Uint) {}

func (x *valClassification) ForBool(q.Bool) {}

func (x *valClassification) ForFloat(q.Float) {}

func (x *valClassification) ForString(q.String) {}

func (x *valClassification) ForUUID(q.UUID) {}

func (x *valClassification) ForBytes(q.Bytes) {}
