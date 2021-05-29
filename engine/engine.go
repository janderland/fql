package engine

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/stream"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

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

func (e *Engine) Set(query keyval.KeyValue) error {
	kind, err := query.Kind()
	if err != nil {
		return errors.Wrap(err, "failed to get query kind")
	}
	if kind != keyval.ConstantKind {
		return errors.New("query not constant kind")
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

		e.log.Debug().Interface("query", query).Msg("setting")

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

		e.log.Debug().Interface("query", query).Msg("clearing")

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

	result, err := e.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dir, err := directory.CreateOrOpen(tr, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open directory")
		}

		e.log.Debug().Interface("query", query).Msg("single reading")

		return tr.Get(dir.Pack(keyval.ToFDBTuple(query.Key.Tuple))).Get()
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	bytes := result.([]byte)
	if bytes == nil {
		return nil, nil
	}

	if len(bytes) > 0 {
		for _, typ := range query.Value.(keyval.Variable) {
			value, err := keyval.UnpackValue(typ, bytes)
			if err != nil {
				continue
			}
			return &keyval.KeyValue{
				Key:   query.Key,
				Value: value,
			}, nil
		}
	}

	return &keyval.KeyValue{
		Key:   query.Key,
		Value: result,
	}, nil
}

func (e *Engine) RangeRead(ctx context.Context, query keyval.KeyValue) chan stream.KeyValErr {
	out := make(chan stream.KeyValErr)

	send := func(msg stream.KeyValErr) {
		select {
		case <-ctx.Done():
		case out <- msg:
		}
	}

	go func() {
		defer close(out)

		kind, err := query.Kind()
		if err != nil {
			send(stream.KeyValErr{Err: errors.Wrap(err, "failed to get query kind")})
			return
		}
		if kind != keyval.RangeReadKind {
			send(stream.KeyValErr{Err: errors.Wrap(err, "query not clear kind")})
			return
		}

		s := stream.New(ctx)
		_, err = e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			stage1 := s.OpenDirectories(tr, query)
			stage2 := s.ReadRange(tr, query, stage1)
			stage3 := s.FilterKeys(query, stage2)
			stage4 := s.UnpackValues(query, stage3)
			for kve := range stage4 {
				send(kve)
			}
			return nil, nil
		})
		if err != nil {
			send(stream.KeyValErr{Err: err})
		}
	}()

	return out
}

func (e *Engine) Directories(ctx context.Context, query keyval.Directory) chan stream.DirErr {
	out := make(chan stream.DirErr)

	send := func(msg stream.DirErr) {
		select {
		case <-ctx.Done():
		case out <- msg:
		}
	}

	go func() {
		defer close(out)

		s := stream.New(ctx)
		_, err := e.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
			for dir := range s.OpenDirectories(tr, keyval.KeyValue{Key: keyval.Key{Directory: query}}) {
				send(dir)
			}
			return nil, nil
		})
		if err != nil {
			send(stream.DirErr{Err: err})
		}
	}()

	return out
}
