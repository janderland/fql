package class

import q "github.com/janderland/fdbq/keyval"

type dirClassification struct{ result subClass }

func (x *dirClassification) VisitString(q.String) {}

func (x *dirClassification) VisitVariable(q.Variable) {
	x.result = variableSubClass
}

type tupClassification struct{ result subClass }

func (x *tupClassification) VisitTuple(e q.Tuple) {
	x.result = classifyTuple(e)
}

func (x *tupClassification) VisitVariable(q.Variable) {
	x.result = variableSubClass
}

func (x *tupClassification) VisitMaybeMore(q.MaybeMore) {
	x.result = variableSubClass
}

func (x *tupClassification) VisitNil(q.Nil) {}

func (x *tupClassification) VisitInt(q.Int) {}

func (x *tupClassification) VisitUint(q.Uint) {}

func (x *tupClassification) VisitBool(q.Bool) {}

func (x *tupClassification) VisitFloat(q.Float) {}

func (x *tupClassification) VisitBigInt(q.BigInt) {}

func (x *tupClassification) VisitString(q.String) {}

func (x *tupClassification) VisitUUID(q.UUID) {}

func (x *tupClassification) VisitBytes(q.Bytes) {}

type valClassification struct{ result subClass }

func (x *valClassification) VisitTuple(e q.Tuple) {
	x.result = classifyTuple(e)
}

func (x *valClassification) VisitVariable(q.Variable) {
	x.result = variableSubClass
}

func (x *valClassification) VisitClear(q.Clear) {
	x.result = clearSubClass
}

func (x *valClassification) VisitNil(q.Nil) {}

func (x *valClassification) VisitInt(q.Int) {}

func (x *valClassification) VisitUint(q.Uint) {}

func (x *valClassification) VisitBool(q.Bool) {}

func (x *valClassification) VisitFloat(q.Float) {}

func (x *valClassification) VisitString(q.String) {}

func (x *valClassification) VisitUUID(q.UUID) {}

func (x *valClassification) VisitBytes(q.Bytes) {}
