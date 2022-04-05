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
	x.str = Directory(key.Directory)
}

func (x *query) ForKeyValue(kv q.KeyValue) {
	x.str = Directory(kv.Key.Directory)
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
