// Package engine provides the code responsible for executing queries.
package engine

import (
	"context"
	"encoding/binary"

	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/engine/internal"
	"github.com/janderland/fdbq/engine/stream"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/keyval/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// SingleOpts specifies how an Engine.SingleRead call is executed.
type SingleOpts struct {
	ByteOrder binary.ByteOrder
	Filter    bool
}

// RangeOpts specifies how an Engine.RangeRead call is executed.
type RangeOpts struct {
	ByteOrder binary.ByteOrder
	Reverse   bool
	Filter    bool
	Limit     int
}

// forStream is used to pass a subset of the options to the
// stream.Stream used in the Engine.RangeRead call.
func (x RangeOpts) forStream() stream.RangeOpts {
	return stream.RangeOpts{
		Reverse: x.Reverse,
		Limit:   x.Limit,
	}
}

// Engine provides methods which execute queries. Each method is built
// for a single class (see package class) of query and will fail if a
// query of the wrong class in provided. Unless Engine.Transact is
// called, each method call is executed in its own transaction.
type Engine struct {
	Tr  facade.Transactor
	Log zerolog.Logger
}

// Transact wraps a group of Engine method calls under a single transaction.
func (e *Engine) Transact(f func(Engine) (interface{}, error)) (interface{}, error) {
	return e.Tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		return f(Engine{Tr: tr, Log: e.Log})
	})
}

// Set preforms a write operation for a single key-value. The given query must
// belong to class.Constant.
func (e *Engine) Set(query q.KeyValue, byteOrder binary.ByteOrder) error {
	if class.Classify(query) != class.Constant {
		return errors.New("query not constant class")
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}

	valueBytes, err := values.Pack(query.Value, byteOrder)
	if err != nil {
		return errors.Wrap(err, "failed to pack value")
	}

	_, err = e.Tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.Log.Log().Interface("query", query).Msg("setting")

		dir, err := tr.DirCreateOrOpen(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert to FDB tuple")
		}

		tr.Set(dir.Pack(tup), valueBytes)
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

// Clear performs a clear operation for a single key-value. The given query
// must belong to class.Clear.
func (e *Engine) Clear(query q.KeyValue) error {
	if class.Classify(query) != class.Clear {
		return errors.New("query not clear class")
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}

	_, err = e.Tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.Log.Log().Interface("query", query).Msg("clearing")

		dir, err := tr.DirOpen(path)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to open directory")
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert to FDB tuple")
		}

		tr.Clear(dir.Pack(tup))
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

// SingleRead performs a read operation for a single key-value. The given query must
// belong to class.SingleRead.
func (e *Engine) SingleRead(query q.KeyValue, opts SingleOpts) (*q.KeyValue, error) {
	if class.Classify(query) != class.SingleRead {
		return nil, errors.New("query not single-read class")
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert directory to string array")
	}

	valHandler, err := internal.NewValueHandler(query.Value, opts.ByteOrder, opts.Filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init value handler")
	}

	var valBytes []byte
	_, err = e.Tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.Log.Log().Interface("query", query).Msg("single reading")

		dir, err := tr.DirOpen(path)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to open directory")
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert to FDB tuple")
		}

		valBytes = tr.Get(dir.Pack(tup)).MustGet()
		return nil, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	if valBytes == nil {
		return nil, nil
	}

	value, err := valHandler.Handle(valBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unpack value")
	}
	if value == nil {
		return nil, nil
	}
	return &q.KeyValue{
		Key:   query.Key,
		Value: value,
	}, nil
}

// RangeRead performs a read across a range of key-values. The given query must belong to class.RangeRead.
// After an error occurs or the entire range is read, the returned channel is closed. If the provided context
// is canceled, then the read operation will be stopped after the current FDB call finishes.
func (e *Engine) RangeRead(ctx context.Context, query q.KeyValue, opts RangeOpts) chan stream.KeyValErr {
	out := make(chan stream.KeyValErr)

	go func() {
		defer close(out)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		s := stream.Stream{Ctx: ctx, Log: e.Log}

		if class.Classify(query) != class.RangeRead {
			s.SendKV(out, stream.KeyValErr{Err: errors.New("query not range-read class")})
			return
		}

		valHandler, err := internal.NewValueHandler(query.Value, opts.ByteOrder, opts.Filter)
		if err != nil {
			s.SendKV(out, stream.KeyValErr{Err: errors.Wrap(err, "failed to init value handler")})
			return
		}

		_, err = e.Tr.ReadTransact(func(tr facade.ReadTransaction) (interface{}, error) {
			stage1 := s.OpenDirectories(tr, query.Key.Directory)
			stage2 := s.ReadRange(tr, query.Key.Tuple, opts.forStream(), stage1)
			stage3 := s.UnpackKeys(query.Key.Tuple, opts.Filter, stage2)
			for kve := range s.UnpackValues(query.Value, valHandler, stage3) {
				s.SendKV(out, kve)
			}
			return nil, nil
		})
		if err != nil {
			s.SendKV(out, stream.KeyValErr{Err: errors.Wrap(err, "transaction failed")})
		}
	}()

	return out
}

// Directories reads directories from the directory layer. If the query contains a keyval.Variable,
// multiple directories may be returned. If the query doesn't contain a keyval.Variable, at most a
// single directory will be returned. After an error occurs or all directories have been read, the
// returned channel is closed. If the provided context is canceled, then the read operation will
// be stopped after the current FDB call finishes.
func (e *Engine) Directories(ctx context.Context, query q.Directory) chan stream.DirErr {
	out := make(chan stream.DirErr)

	go func() {
		defer close(out)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		s := stream.Stream{Ctx: ctx, Log: e.Log}

		_, err := e.Tr.ReadTransact(func(tr facade.ReadTransaction) (interface{}, error) {
			for dir := range s.OpenDirectories(tr, query) {
				s.SendDir(out, dir)
			}
			return nil, nil
		})
		if err != nil {
			s.SendDir(out, stream.DirErr{Err: errors.Wrap(err, "transaction failed")})
		}
	}()

	return out
}
