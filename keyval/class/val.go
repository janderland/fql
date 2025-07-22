package class

import q "github.com/janderland/fql/keyval"

func getAttributesOfVal(val q.Value) attributes {
	var attr valAttributes
	if val == nil {
		return attributes{hasNil: true}
	}
	val.Value(&attr)
	return attr.merge(attributes{})
}

var _ q.ValueOperation = &valAttributes{}

type valAttributes struct {
	vstampFutures int
	hasVariable   bool
	hasClear      bool
}

func (x *valAttributes) merge(c attributes) attributes {
	c.vstampFutures = c.vstampFutures + x.vstampFutures
	c.hasVariable = c.hasVariable || x.hasVariable
	c.hasClear = c.hasClear || x.hasClear
	return c
}

func (x *valAttributes) ForTuple(e q.Tuple) {
	subAttr := getAttributesOfTup(e)
	x.vstampFutures = x.vstampFutures + subAttr.vstampFutures
	x.hasVariable = x.hasVariable || subAttr.hasVariable
}

func (x *valAttributes) ForVariable(q.Variable) {
	x.hasVariable = true
}

func (x *valAttributes) ForClear(q.Clear) {
	x.hasClear = true
}

func (x *valAttributes) ForVStampFuture(e q.VStampFuture) {
	x.vstampFutures++
}

func (x *valAttributes) ForNil(q.Nil) {}

func (x *valAttributes) ForInt(q.Int) {}

func (x *valAttributes) ForUint(q.Uint) {}

func (x *valAttributes) ForBool(q.Bool) {}

func (x *valAttributes) ForFloat(q.Float) {}

func (x *valAttributes) ForString(q.String) {}

func (x *valAttributes) ForUUID(q.UUID) {}

func (x *valAttributes) ForBytes(q.Bytes) {}

func (x *valAttributes) ForVStamp(q.VStamp) {}
