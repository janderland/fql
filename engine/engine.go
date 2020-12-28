package engine

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type Engine struct {
	DB fdb.Transactor
}

func (e *Engine) Execute(queries []keyval.KeyValue) ([]interface{}, error) {
	results, err := e.DB.Transact(func(tr fdb.Transaction) (interface{}, error) {
		var results []interface{}
		for i, q := range queries {
			if _, ok := q.Value.(keyval.Clear); ok {
				err := e.clear(q)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to execute query %d as clear", i)
				}
			}
			if queryHasVariable(q) {
				res, err := e.get(q)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to execute query %d as get", i)
				}
				results = append(results, res)
			}
			err := e.set(q)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to execute query %d as set", i)
			}
		}
		return results, nil
	})
	if results != nil {
		return results.([]interface{}), err
	}
	return nil, err
}

func (e *Engine) clear(_ keyval.KeyValue) error {
	return errors.New("not implemented")
}

func (e *Engine) get(_ keyval.KeyValue) (interface{}, error) {
	return nil, nil
}

func (e *Engine) set(_ keyval.KeyValue) error {
	return errors.New("not implemented")
}

func queryHasVariable(q keyval.KeyValue) bool {
	for _, dir := range q.Key.Directory {
		if _, ok := dir.(keyval.Variable); ok {
			return true
		}
	}

	if tupleHasVariable(q.Key.Tuple) {
		return true
	}

	switch q.Value.(type) {
	case keyval.Tuple:
		if tupleHasVariable(q.Value.(keyval.Tuple)) {
			return true
		}
	case keyval.Variable:
		return true
	}

	return false
}

func tupleHasVariable(tuple keyval.Tuple) bool {
	for _, element := range tuple {
		switch element.(type) {
		case keyval.Tuple:
			if tupleHasVariable(element.(keyval.Tuple)) {
				return true
			}
		case keyval.Variable:
			return true
		}
	}
	return false
}
