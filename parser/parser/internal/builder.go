package internal

import q "github.com/janderland/fdbq/keyval"

// KeyValBuilder is used by parser.Parser to construct the
// resultant key-value. parser.Parser doesn't interact with
// keyval.KeyValue directly, so these methods outline all
// the key-value state changes performed by the parser.
type KeyValBuilder struct {
	kv q.KeyValue
}

func (x *KeyValBuilder) Get() q.KeyValue {
	return x.kv
}

func (x *KeyValBuilder) AppendVarToDirectory() {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.Variable{})
}

func (x *KeyValBuilder) AppendPartToDirectory(token string) {
	x.kv.Key.Directory = append(x.kv.Key.Directory, q.String(token))
}

func (x *KeyValBuilder) AppendToLastDirPart(token string) {
	i := len(x.kv.Key.Directory) - 1
	str := x.kv.Key.Directory[i].(q.String)
	x.kv.Key.Directory[i] = q.String(string(str) + token)
}

func (x *KeyValBuilder) AppendToValueVar(typ q.ValueType) {
	x.kv.Value = append(x.kv.Value.(q.Variable), typ)
}

func (x *KeyValBuilder) SetKeyTuple(tup q.Tuple) {
	x.kv.Key.Tuple = tup
}

func (x *KeyValBuilder) SetValue(val q.Value) {
	x.kv.Value = val
}

// TupBuilder is used by parser.Parser to construct tuples.
// parser.Parser doesn't interact with keyval.Tuple directly,
// so these methods outline all the key-value state changes
// performed by the parser.
type TupBuilder struct {
	root  q.Tuple
	depth int
}

func (x *TupBuilder) Get() q.Tuple {
	return x.root
}

func (x *TupBuilder) StartSubTuple() {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, q.Tuple{})
	})
	x.depth++
}

func (x *TupBuilder) EndTuple() bool {
	x.depth--
	return x.depth == -1
}

func (x *TupBuilder) Append(e q.TupElement) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, e)
	})
}

func (x *TupBuilder) AppendToLastElemStr(token string) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		str := tup[i].(q.String)
		tup[i] = q.String(string(str) + token)
		return tup
	})
}

func (x *TupBuilder) AppendToLastElemVar(typ q.ValueType) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		v := tup[i].(q.Variable)
		tup[i] = append(v, typ)
		return tup
	})
}

func (x *TupBuilder) mutateTuple(f func(q.Tuple) q.Tuple) {
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
