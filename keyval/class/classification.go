package class

import q "github.com/janderland/fdbq/keyval"

var (
	_ q.DirectoryOperation = &dirClassification{}
	_ q.TupleOperation     = &tupClassification{}
	_ q.ValueOperation     = &valClassification{}
)

type dirClassification struct{ result subClass }

func (x *dirClassification) ForString(q.String) {}

func (x *dirClassification) ForVariable(q.Variable) {
	x.result = variableSubClass
}

type tupClassification struct{ result subClass }

func (x *tupClassification) ForTuple(e q.Tuple) {
	x.result = classifyTuple(e)
}

func (x *tupClassification) ForVariable(q.Variable) {
	x.result = variableSubClass
}

func (x *tupClassification) ForMaybeMore(q.MaybeMore) {
	x.result = variableSubClass
}

func (x *tupClassification) ForNil(q.Nil) {}

func (x *tupClassification) ForInt(q.Int) {}

func (x *tupClassification) ForUint(q.Uint) {}

func (x *tupClassification) ForBool(q.Bool) {}

func (x *tupClassification) ForFloat(q.Float) {}

func (x *tupClassification) ForBigInt(q.BigInt) {}

func (x *tupClassification) ForString(q.String) {}

func (x *tupClassification) ForUUID(q.UUID) {}

func (x *tupClassification) ForBytes(q.Bytes) {}

type valClassification struct{ result subClass }

func (x *valClassification) ForTuple(e q.Tuple) {
	x.result = classifyTuple(e)
}

func (x *valClassification) ForVariable(q.Variable) {
	x.result = variableSubClass
}

func (x *valClassification) ForClear(q.Clear) {
	x.result = clearSubClass
}

func (x *valClassification) ForNil(q.Nil) {}

func (x *valClassification) ForInt(q.Int) {}

func (x *valClassification) ForUint(q.Uint) {}

func (x *valClassification) ForBool(q.Bool) {}

func (x *valClassification) ForFloat(q.Float) {}

func (x *valClassification) ForString(q.String) {}

func (x *valClassification) ForUUID(q.UUID) {}

func (x *valClassification) ForBytes(q.Bytes) {}
