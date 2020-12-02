package engine

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/query"
	"github.com/pkg/errors"
)

type Engine struct {
	DB fdb.Transactor
}

func (e *Engine) Execute(queries []query.Query) ([]interface{}, error) {
	results, err := e.DB.Transact(func(tr fdb.Transaction) (interface{}, error) {
		var results []interface{}
		for i, q := range queries {
			if _, ok := q.Value.(query.Clear); ok {
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

func (e *Engine) clear(_ query.Query) error {
	return errors.New("not implemented")
}

func (e *Engine) get(_ query.Query) (interface{}, error) {
	return nil, nil
}

func (e *Engine) set(_ query.Query) error {
	return errors.New("not implemented")
}

func queryHasVariable(q query.Query) bool {
	for _, dir := range q.Key.Directory {
		if _, ok := dir.(query.Variable); ok {
			return true
		}
	}

	if tupleHasVariable(q.Key.Tuple) {
		return true
	}

	switch q.Value.(type) {
	case query.Tuple:
		if tupleHasVariable(q.Value.(query.Tuple)) {
			return true
		}
	case query.Variable:
		return true
	}

	return false
}

func tupleHasVariable(tuple query.Tuple) bool {
	for _, element := range tuple {
		switch element.(type) {
		case query.Tuple:
			if tupleHasVariable(element.(query.Tuple)) {
				return true
			}
		case query.Variable:
			return true
		}
	}
	return false
}
