package parser

import q "github.com/janderland/fdbq/keyval"

type kvBuilder struct {
	q.KeyValue
	tupBuilder
}

func (x *kvBuilder) get() q.KeyValue {
	return x.KeyValue
}

func (x *kvBuilder) appendVarToDirectory() {
	x.Key.Directory = append(x.Key.Directory, q.Variable{})
}

func (x *kvBuilder) appendPartToDirectory(token string) {
	x.Key.Directory = append(x.Key.Directory, q.String(token))
}

func (x *kvBuilder) appendToLastDirPart(token string) {
	i := len(x.Key.Directory) - 1
	str := x.Key.Directory[i].(q.String)
	x.Key.Directory[i] = q.String(string(str) + token)
}

func (x *kvBuilder) startTuple() {
	x.tupBuilder = tupBuilder{}
}

func (x *kvBuilder) endKeyTuple() {
	x.Key.Tuple = x.tupBuilder.get()
}

func (x *kvBuilder) endValueTuple() {
	x.Value = x.tupBuilder.get()
}

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

func (x *tupBuilder) endCurrentTuple() bool {
	x.depth--
	if x.depth == -1 {
		return true
	}
	return false
}

func (x *tupBuilder) appendStringToTuple() {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		return append(tup, q.String(""))
	})
}

func (x *tupBuilder) appendToLastTupElem(token string) {
	x.mutateTuple(func(tup q.Tuple) q.Tuple {
		i := len(tup) - 1
		str := tup[i].(q.String)
		tup[i] = q.String(string(str) + token)
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
