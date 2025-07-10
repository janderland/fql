package class

import q "github.com/janderland/fql/keyval"

func getAttributesOfTup(tup q.Tuple) attributes {
	var (
		attr   tupAttributes
		hasNil bool
	)
	for _, element := range tup {
		if element == nil {
			hasNil = true
			continue
		}
		element.TupElement(&attr)
	}
	return attr.merge(attributes{hasNil: hasNil})
}

var _ q.TupleOperation = &tupAttributes{}

type tupAttributes struct {
	vstampFutures int
	hasVariable   bool
	hasNil        bool
}

func (x *tupAttributes) merge(c attributes) attributes {
	c.vstampFutures = c.vstampFutures + x.vstampFutures
	c.hasVariable = c.hasVariable || x.hasVariable
	c.hasNil = c.hasNil || x.hasNil
	return c
}

func (x *tupAttributes) ForTuple(e q.Tuple) {
	class := getAttributesOfTup(e)
	x.vstampFutures = x.vstampFutures + class.vstampFutures
	x.hasVariable = x.hasVariable || class.hasVariable
	x.hasNil = x.hasNil || class.hasNil
}

func (x *tupAttributes) ForVariable(q.Variable) {
	x.hasVariable = true
}

func (x *tupAttributes) ForMaybeMore(q.MaybeMore) {
	x.hasVariable = true
}

func (x *tupAttributes) ForNil(q.Nil) {}

func (x *tupAttributes) ForInt(q.Int) {}

func (x *tupAttributes) ForUint(q.Uint) {}

func (x *tupAttributes) ForBool(q.Bool) {}

func (x *tupAttributes) ForFloat(q.Float) {}

func (x *tupAttributes) ForString(q.String) {}

func (x *tupAttributes) ForUUID(q.UUID) {}

func (x *tupAttributes) ForBytes(q.Bytes) {}

func (x *tupAttributes) ForVStamp(q.VStamp) {}

func (x *tupAttributes) ForVStampFuture(q.VStampFuture) {
	x.vstampFutures++
}
