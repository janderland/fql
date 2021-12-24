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

type SingleOpts struct {
	ByteOrder binary.ByteOrder
	Filter    bool
}

type RangeOpts struct {
	ByteOrder binary.ByteOrder
	Reverse   bool
	Filter    bool
	Limit     int
}

func (x RangeOpts) forStream() stream.RangeOpts {
	return stream.RangeOpts{
		Reverse: x.Reverse,
		Limit:   x.Limit,
	}
}

type Engine struct {
	tr  facade.Transactor
	ctx context.Context
	log *zerolog.Logger
}

func New(ctx context.Context, tr facade.Transactor) Engine {
	return Engine{
		tr:  tr,
		ctx: ctx,
		log: zerolog.Ctx(ctx),
	}
}

func (e *Engine) Transact(f func(Engine) (interface{}, error)) (interface{}, error) {
	return e.tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		return f(New(e.ctx, tr))
	})
}

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

	_, err = e.tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("setting")

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

func (e *Engine) Clear(query q.KeyValue) error {
	if class.Classify(query) != class.Clear {
		return errors.New("query not clear class")
	}

	path, err := convert.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}

	_, err = e.tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("clearing")

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
	_, err = e.tr.Transact(func(tr facade.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("single reading")

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

func (e *Engine) RangeRead(ctx context.Context, query q.KeyValue, opts RangeOpts) chan stream.KeyValErr {
	out := make(chan stream.KeyValErr)

	go func() {
		defer close(out)

		s, stop := stream.New(ctx)
		defer stop()

		if class.Classify(query) != class.RangeRead {
			s.SendKV(out, stream.KeyValErr{Err: errors.New("query not range-read class")})
			return
		}

		valHandler, err := internal.NewValueHandler(query.Value, opts.ByteOrder, opts.Filter)
		if err != nil {
			s.SendKV(out, stream.KeyValErr{Err: errors.Wrap(err, "failed to init value handler")})
			return
		}

		_, err = e.tr.ReadTransact(func(tr facade.ReadTransaction) (interface{}, error) {
			stage1 := s.OpenDirectories(tr, query.Key.Directory)
			stage2 := s.ReadRange(tr, query.Key.Tuple, opts.forStream(), stage1)
			stage3 := s.FilterKeys(query.Key.Tuple, opts.Filter, stage2)
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

func (e *Engine) Directories(ctx context.Context, query q.Directory) chan stream.DirErr {
	out := make(chan stream.DirErr)

	go func() {
		defer close(out)

		s, stop := stream.New(ctx)
		defer stop()

		_, err := e.tr.ReadTransact(func(tr facade.ReadTransaction) (interface{}, error) {
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
