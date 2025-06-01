package class

import q "github.com/janderland/fql/keyval"

func getAttributesOfDir(dir q.Directory) attributes {
	var (
		attr   dirAttributes
		hasNil bool
	)
	for _, element := range dir {
		if element == nil {
			hasNil = true
			continue
		}
		element.DirElement(&attr)
	}
	return attr.merge(attributes{hasNil: hasNil})
}

var _ q.DirectoryOperation = &dirAttributes{}

type dirAttributes struct {
	hasVariable bool
}

func (x *dirAttributes) merge(c attributes) attributes {
	c.hasVariable = c.hasVariable || x.hasVariable
	return c
}

func (x *dirAttributes) ForString(q.String) {}

func (x *dirAttributes) ForVariable(q.Variable) {
	x.hasVariable = true
}
