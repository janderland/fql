package internal

import q "github.com/janderland/fdbq/keyval"

// KVBuilder is used by Parser to manipulate a keyval.KeyValue.
// Parser doesn't interact with keyval.KeyValue directly, so
// these methods outline all the keyval.KeyValue state changes
// performed by the Parser.
type KVBuilder struct {
	kv q.KeyValue
}

func (x *KVBuilder) Get() q.KeyValue {
	return x.kv
}

func (x *KVBuilder) AppendVarToDirectory() {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.Variable{})
}

func (x *KVBuilder) AppendPartToDirectory(token string) {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.String(token))
}

func (x *KVBuilder) AppendToLastDirPart(token string) {
	i := len(x.kv.Key.Directory) - 1
	str := x.kv.Key.Directory[i].(q.String)
	x.kv.Key.Directory[i] = q.String(string(str) + token)
}

func (x *KVBuilder) AppendToValueVar(typ q.ValueType) {
	x.kv.Value = append(x.kv.Value.(q.Variable), typ)
}

func (x *KVBuilder) SetKeyTuple(tup q.Tuple) {
	x.kv.Key.Tuple = tup
}

func (x *KVBuilder) SetValue(val q.Value) {
	x.kv.Value = val
}

// TupBuilder is used by Parser to manipulate a keyval.Tuple.
// Parser doesn't interact with keyval.Tuple directly, so
// these methods outline all the keyval.Tuple state changes
// performed by the Parser.
type TupBuilder struct {
	root  q.Tuple
	depth int
}

func (x *TupBuilder) Get() q.Tuple {
	return x.root
}

func (x *TupBuilder) StartSubTuple() {
	x.MutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, q.Tuple{})
	})
	x.depth++
}

func (x *TupBuilder) EndTuple() bool {
	x.depth--
	return x.depth == -1
}

func (x *TupBuilder) Append(e q.TupElement) {
	x.MutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, e)
	})
}

func (x *TupBuilder) AppendToLastElemStr(token string) {
	x.MutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		str := tup[i].(q.String)
		tup[i] = q.String(string(str) + token)
		return tup
	})
}

func (x *TupBuilder) AppendToLastElemVar(typ q.ValueType) {
	x.MutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		v := tup[i].(q.Variable)
		tup[i] = append(v, typ)
		return tup
	})
}

func (x *TupBuilder) MutateTuple(f func(q.Tuple) q.Tuple) {
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
