package engine

import (
	"context"
	"encoding/binary"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/stream"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Engine struct {
	db    fdb.Transactor
	ctx   context.Context
	log   *zerolog.Logger
	order binary.ByteOrder
}

func New(ctx context.Context, db fdb.Transactor, order binary.ByteOrder) Engine {
	return Engine{
		db:    db,
		ctx:   ctx,
		log:   zerolog.Ctx(ctx),
		order: order,
	}
}

func (e *Engine) Transact(f func(Engine) (interface{}, error)) (interface{}, error) {
	return e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(New(e.ctx, tr, e.order))
	})
}

func (e *Engine) Set(query q.KeyValue) error {
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get query kind")
	}
	if kind != q.ConstantKind {
		return errors.New("query not constant kind")
	}

	path, err := q.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}
	valueBytes, err := q.PackValue(e.order, query.Value)
	if err != nil {
		return errors.Wrap(err, "failed to pack value")
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("setting")

		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}
		tr.Set(dir.Pack(q.ToFDBTuple(query.Key.Tuple)), valueBytes)
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

func (e *Engine) Clear(query q.KeyValue) error {
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get query kind")
	}
	if kind != q.ClearKind {
		return errors.New("query not clear kind")
	}

	path, err := q.ToStringArray(query.Key.Directory)
	if err != nil {
		return errors.Wrap(err, "failed to convert directory to string array")
	}

	_, err = e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("clearing")

		dir, err := directory.Open(tr, path, nil)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to open directory")
		}
		tr.Clear(dir.Pack(q.ToFDBTuple(query.Key.Tuple)))
		return nil, nil
	})
	return errors.Wrap(err, "transaction failed")
}

func (e *Engine) SingleRead(query q.KeyValue) (*q.KeyValue, error) {
	kind, err := query.Kind()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get query kind")
	}
	if kind != q.SingleReadKind {
		return nil, errors.New("query not single-read kind")
	}

	path, err := q.ToStringArray(query.Key.Directory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert directory to string array")
	}
	unpack, err := q.NewUnpack(query.Value, e.order)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init unpacker")
	}

	result, err := e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		e.log.Log().Interface("query", query).Msg("single reading")

		dir, err := directory.Open(tr, path, nil)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to open directory")
		}
		return tr.Get(dir.Pack(q.ToFDBTuple(query.Key.Tuple))).Get()
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

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

func (e *Engine) RangeRead(ctx context.Context, query q.KeyValue, opts fdb.RangeOptions) chan stream.KeyValErr {
	out := make(chan stream.KeyValErr)

	go func() {
		defer close(out)

		s, stop := stream.New(ctx)
		defer stop()

		kind, err := query.Kind()
		if err != nil {
			s.SendKV(out, stream.KeyValErr{Err: errors.Wrap(err, "failed to get query kind")})
			return
		}
		if kind != q.RangeReadKind {
			s.SendKV(out, stream.KeyValErr{Err: errors.New("query not range-read kind")})
			return
		}

		_, err = e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			stage1 := s.OpenDirectories(tr, query.Key.Directory)
			stage2 := s.ReadRange(tr, query.Key.Tuple, opts, stage1)
			stage3 := s.FilterKeys(query.Key.Tuple, stage2)
			for kve := range s.UnpackValues(query.Value, e.order, stage3) {
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

		_, err := e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
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
