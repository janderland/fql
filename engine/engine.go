package engine

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/kv"
	"github.com/pkg/errors"
)

type Engine struct {
	DB fdb.Transactor
}

func (e *Engine) Execute(queries []kv.KeyValue) ([]interface{}, error) {
	results, err := e.DB.Transact(func(tr fdb.Transaction) (interface{}, error) {
		var results []interface{}
		for i, q := range queries {
			if _, ok := q.Value.(kv.Clear); ok {
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

func (e *Engine) clear(_ kv.KeyValue) error {
	return errors.New("not implemented")
}

func (e *Engine) get(_ kv.KeyValue) (interface{}, error) {
	return nil, nil
}

func (e *Engine) set(_ kv.KeyValue) error {
	return errors.New("not implemented")
}

func queryHasVariable(q kv.KeyValue) bool {
	for _, dir := range q.Key.Directory {
		if _, ok := dir.(kv.Variable); ok {
			return true
		}
	}

	if tupleHasVariable(q.Key.Tuple) {
		return true
	}

	switch q.Value.(type) {
	case kv.Tuple:
		if tupleHasVariable(q.Value.(kv.Tuple)) {
			return true
		}
	case kv.Variable:
		return true
	}

	return false
}

func tupleHasVariable(tuple kv.Tuple) bool {
	for _, element := range tuple {
		switch element.(type) {
		case kv.Tuple:
			if tupleHasVariable(element.(kv.Tuple)) {
				return true
			}
		case kv.Variable:
			return true
		}
	}
	return false
}
