package engine

import (
	"context"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/reader"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type Engine struct{ db fdb.Transactor }

func (e *Engine) Transact(f func(Engine) (interface{}, error)) (interface{}, error) {
	return e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(Engine{tr})
	})
}

func (e *Engine) Set(query keyval.KeyValue) error {
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get query kind")
	}
	if kind != keyval.ConstantKind {
		return errors.Wrap(err, "query not constant kind")
	}

	path, err := keyval.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}
	valueBytes, err := keyval.PackValue(query.Value)
	if err != nil {
		return errors.Wrap(err, "failed to pack value")
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}
		tr.Set(dir.Pack(keyval.ToFDBTuple(query.Key.Tuple)), valueBytes)
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

func (e *Engine) Clear(query keyval.KeyValue) error {
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get query kind")
	}
	if kind != keyval.ClearKind {
		return errors.Wrap(err, "query not clear kind")
	}

	path, err := keyval.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}
		tr.Clear(dir.Pack(keyval.ToFDBTuple(query.Key.Tuple)))
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

func (e *Engine) SingleRead(query keyval.KeyValue) (*keyval.KeyValue, error) {
	kind, err := query.Kind()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get query kind")
	}
	if kind != keyval.SingleReadKind {
		return nil, errors.Wrap(err, "query not clear kind")
	}

	path, err := keyval.ToStringArray(query.Key.Directory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert directory to string array")
	}

	bytes, err := e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}
		return tr.Get(dir.Pack(keyval.ToFDBTuple(query.Key.Tuple))).Get()
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	for _, typ := range query.Value.(keyval.Variable) {
		value, err := keyval.UnpackValue(typ, bytes.([]byte))
		if err != nil {
			continue
		}
		return &keyval.KeyValue{
			Key:   query.Key,
			Value: value,
		}, nil
	}
	return nil, nil
}

type KeyValErr struct {
	KV  keyval.KeyValue
	Err error
}

func (e *Engine) RangeRead(ctx context.Context, query keyval.KeyValue) chan KeyValErr {
	out := make(chan KeyValErr)

	send := func(msg KeyValErr) {
		select {
		case <-ctx.Done():
		case out <- msg:
		}
	}

	go func() {
		defer close(out)

		kind, err := query.Kind()
		if err != nil {
			send(KeyValErr{Err: errors.Wrap(err, "failed to get query kind")})
			return
		}
		if kind != keyval.RangeReadKind {
			send(KeyValErr{Err: errors.Wrap(err, "query not clear kind")})
			return
		}

		_, err = e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			r := reader.New(ctx, tr)
			kvCh, errCh := r.Read(query)

			for {
				select {
				case err, open := <-errCh:
					if !open {
						return nil, nil
					}
					return nil, err

				case kv, open := <-kvCh:
					if !open {
						return nil, nil
					}
					send(KeyValErr{KV: kv})
				}
			}
		})
		if err != nil {
			send(KeyValErr{Err: err})
		}
	}()

	return out
}
