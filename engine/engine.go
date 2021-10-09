package engine

import (
	"context"
	"encoding/binary"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/stream"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/class"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/keyval/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	ClassNotConstantErr   = "query not constant class"
	ClassNotClearErr      = "query not clear class"
	ClassNotSingleReadErr = "query not single-read class"
	ClassNotRangeReadErr  = "query not range-read class"

	StringArrayErr   = "failed to convert directory to string array"
	TupleErr         = "failed to convert to FDB tuple"
	OpenDirectoryErr = "failed to open directory"
	TransactionErr   = "transaction failed"

	PackValueErr   = "failed to pack value"
	NewUnpackerErr = "failed to init unpacker"
)

type RangeOpts struct {
	ByteOrder binary.ByteOrder
	Reverse   bool
	Limit     int
}

func (x RangeOpts) forStream() stream.RangeOpts {
	return stream.RangeOpts{
		Reverse: x.Reverse,
		Limit:   x.Limit,
	}
}

type Engine struct {
	db  fdb.Transactor
	ctx context.Context
	log *zerolog.Logger
}

func New(ctx context.Context, db fdb.Transactor) Engine {
	return Engine{
		db:  db,
		ctx: ctx,
		log: zerolog.Ctx(ctx),
	}
}

func (e *Engine) Transact(f func(Engine) (interface{}, error)) (interface{}, error) {
	return e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(New(e.ctx, tr))
	})
}

func (e *Engine) Set(query q.KeyValue, byteOrder binary.ByteOrder) error {
	if class.Classify(query) != class.Constant {
		return errors.New(ClassNotConstantErr)
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, StringArrayErr)
	}

	valueBytes, err := values.PackValue(query.Value, byteOrder)
	if err != nil {
		return errors.Wrap(err, PackValueErr)
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("setting")

		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, OpenDirectoryErr)
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, TupleErr)
		}

		tr.Set(dir.Pack(tup), valueBytes)
		return nil, nil
	})
	return errors.Wrap(err, TransactionErr)
}

func (e *Engine) Clear(query q.KeyValue) error {
	if class.Classify(query) != class.Clear {
		return errors.New(ClassNotClearErr)
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, StringArrayErr)
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("clearing")

		dir, err := directory.Open(tr, path, nil)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, OpenDirectoryErr)
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, TupleErr)
		}

		tr.Clear(dir.Pack(tup))
		return nil, nil
	})
	return errors.Wrap(err, TransactionErr)
}

func (e *Engine) SingleRead(query q.KeyValue, byteOrder binary.ByteOrder) (*q.KeyValue, error) {
	if class.Classify(query) != class.SingleRead {
		return nil, errors.New(ClassNotSingleReadErr)
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return nil, errors.Wrap(err, StringArrayErr)
	}

	unpack, err := values.NewUnpack(query.Value, byteOrder)
	if err != nil {
		return nil, errors.Wrap(err, NewUnpackerErr)
	}

	result, err := e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("single reading")

		dir, err := directory.Open(tr, path, nil)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, OpenDirectoryErr)
		}

		tup, err := convert.ToFDBTuple(query.Key.Tuple)
		if err != nil {
			return nil, errors.Wrap(err, TupleErr)
		}

		return tr.Get(dir.Pack(tup)).Get()
	})
	if err != nil {
		return nil, errors.Wrap(err, TransactionErr)
	}

	// Before asserting the result is a []byte, we need
	// to check if it's an untyped nil. This check could
	// be avoided if we ensured the transaction callback
	// always returned a []byte, but having this check
	// here is less fragile.
	if result == nil {
		return nil, nil
	}

	bytes := result.([]byte)
	if bytes == nil {
		return nil, nil
	}
	value := unpack(bytes)
	if value == nil {
		return nil, nil
	}
	return &q.KeyValue{
		Key:   query.Key,
		Value: value,
	}, nil
}

func (e *Engine) RangeRead(ctx context.Context, query q.KeyValue, opts RangeOpts) chan stream.KeyValErr {
	out := make(chan stream.KeyValErr)

	go func() {
		defer close(out)

		s, stop := stream.New(ctx)
		defer stop()

		if class.Classify(query) != class.RangeRead {
			s.SendKV(out, stream.KeyValErr{Err: errors.New(ClassNotRangeReadErr)})
			return
		}

		_, err := e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			stage1 := s.OpenDirectories(tr, query.Key.Directory)
			stage2 := s.ReadRange(tr, query.Key.Tuple, opts.forStream(), stage1)
			stage3 := s.FilterKeys(query.Key.Tuple, stage2)
			for kve := range s.UnpackValues(query.Value, opts.ByteOrder, stage3) {
				s.SendKV(out, kve)
			}
			return nil, nil
		})
		if err != nil {
			s.SendKV(out, stream.KeyValErr{Err: errors.Wrap(err, TransactionErr)})
		}
	}()

	return out
}

func (e *Engine) Directories(ctx context.Context, query q.Directory) chan stream.DirErr {
	out := make(chan stream.DirErr)

	go func() {
		defer close(out)

		s, stop := stream.New(ctx)
		defer stop()

		_, err := e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			for dir := range s.OpenDirectories(tr, query) {
				s.SendDir(out, dir)
			}
			return nil, nil
		})
		if err != nil {
			s.SendDir(out, stream.DirErr{Err: errors.Wrap(err, TransactionErr)})
		}
	}()

	return out
}
