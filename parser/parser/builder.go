package parser

import q "github.com/janderland/fdbq/keyval"

// kvBuilder is used by Parser to manipulate a keyval.KeyValue.
// Parser doesn't interact with keyval.KeyValue directly, so
// these methods outline all the keyval.KeyValue state changes
// performed by the Parser.
type kvBuilder struct {
	kv q.KeyValue
}

func (x *kvBuilder) get() q.KeyValue {
	return x.kv
}

func (x *kvBuilder) appendVarToDirectory() {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.Variable{})
}

func (x *kvBuilder) appendPartToDirectory(token string) {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.String(token))
}

func (x *kvBuilder) appendToLastDirPart(token string) {
	i := len(x.kv.Key.Directory) - 1
	str := x.kv.Key.Directory[i].(q.String)
	x.kv.Key.Directory[i] = q.String(string(str) + token)
}

func (x *kvBuilder) appendToValueVar(typ q.ValueType) {
	x.kv.Value = append(x.kv.Value.(q.Variable), typ)
}

func (x *kvBuilder) setKeyTuple(tup q.Tuple) {
	x.kv.Key.Tuple = tup
}

func (x *kvBuilder) setValue(val q.Value) {
	x.kv.Value = val
}

// tupBuilder is used by Parser to manipulate a keyval.Tuple.
// Parser doesn't interact with keyval.Tuple directly, so
// these methods outline all the keyval.Tuple state changes
// performed by the Parser.
type tupBuilder struct {
	root  q.Tuple
	depth int
}

func (x *tupBuilder) get() q.Tuple {
	return x.root
}

func (x *tupBuilder) startSubTuple() {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, q.Tuple{})
	})
	x.depth++
}

func (x *tupBuilder) endTuple() bool {
	x.depth--
	return x.depth == -1
}

func (x *tupBuilder) append(e q.TupElement) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, e)
	})
}

func (x *tupBuilder) appendToLastElemStr(token string) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		str := tup[i].(q.String)
		tup[i] = q.String(string(str) + token)
		return tup
	})
}

func (x *tupBuilder) appendToLastElemVar(typ q.ValueType) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		v := tup[i].(q.Variable)
		tup[i] = append(v, typ)
		return tup
	})
}

func (x *tupBuilder) mutateTuple(f func(q.Tuple) q.Tuple) {
	tuples := []q.Tuple{x.root}
	tup := tuples[0]

	for i := 0; i < x.depth; i++ {
		tup = tup[len(tup)-1].(q.Tuple)
		tuples = append(tuples, tup)
	}

	tup = f(tup)
	tuples[len(tuples)-1] = tup

	for i := len(tuples) - 1; i > 0; i-- {
		tuples[i-1][len(tuples[i-1])-1] = tuples[i]
	}

	x.root = tuples[0]
}
