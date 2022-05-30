package stream

import (
	"context"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/engine/internal"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/compare"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type (
	RangeOpts struct {
		Reverse bool
		Limit   int
	}

	Stream struct {
		Ctx context.Context
		Log zerolog.Logger
	}

	DirErr struct {
		Dir directory.DirectorySubspace
		Err error
	}

	DirKVErr struct {
		Dir directory.DirectorySubspace
		KV  fdb.KeyValue
		Err error
	}

	KeyValErr struct {
		KV  q.KeyValue
		Err error
	}
)

func (r *Stream) SendDir(out chan<- DirErr, in DirErr) bool {
	select {
	case <-r.Ctx.Done():
		return false
	case out <- in:
		return true
	}
}

func (r *Stream) SendDirKV(out chan<- DirKVErr, in DirKVErr) bool {
	select {
	case <-r.Ctx.Done():
		return false
	case out <- in:
		return true
	}
}

func (r *Stream) SendKV(out chan<- KeyValErr, in KeyValErr) bool {
	select {
	case <-r.Ctx.Done():
		return false
	case out <- in:
		return true
	}
}

func (r *Stream) OpenDirectories(tr facade.ReadTransactor, query q.Directory) chan DirErr {
	out := make(chan DirErr)

	go func() {
		defer close(out)
		r.goOpenDirectories(tr, query, out)
	}()

	return out
}

func (r *Stream) ReadRange(tr facade.ReadTransaction, query q.Tuple, opts RangeOpts, in chan DirErr) chan DirKVErr {
	out := make(chan DirKVErr)

	go func() {
		defer close(out)
		r.goReadRange(tr, query, opts, in, out)
	}()

	return out
}

func (r *Stream) UnpackKeys(query q.Tuple, filter bool, in chan DirKVErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		r.goUnpackKeys(query, filter, in, out)
	}()

	return out
}

func (r *Stream) UnpackValues(query q.Value, valHandler internal.ValHandler, in chan KeyValErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		r.goUnpackValues(query, valHandler, in, out)
	}()

	return out
}

func (r *Stream) goOpenDirectories(tr facade.ReadTransactor, query q.Directory, out chan DirErr) {
	log := r.Log.With().Str("stage", "open directories").Interface("query", query).Logger()

	prefix, variable, suffix := splitAtFirstVariable(query)
	prefixStr, err := convert.ToStringArray(prefix)
	if err != nil {
		r.SendDir(out, DirErr{Err: errors.Wrapf(err, "failed to convert directory prefix to string array")})
		return
	}

	if variable != nil {
		subDirs, err := tr.DirList(prefixStr)
		if err != nil {
			r.SendDir(out, DirErr{Err: errors.Wrap(err, "failed to list directories")})
			return
		}
		if len(subDirs) == 0 {
			log.Log().Msg("no subdirectories")
			return
		}

		log.Log().Strs("sub dirs", subDirs).Msg("found subdirectories")

		for _, subDir := range subDirs {
			// Between each interaction with the DB, give
			// this goroutine a chance to exit early.
			if err := r.Ctx.Err(); err != nil {
				r.SendDir(out, DirErr{Err: err})
				return
			}

			var dir q.Directory
			dir = append(dir, prefix...)
			dir = append(dir, q.String(subDir))
			dir = append(dir, suffix...)
			r.goOpenDirectories(tr, dir, out)
		}
	} else {
		dir, err := tr.DirOpen(prefixStr)
		if err != nil {
			r.SendDir(out, DirErr{Err: errors.Wrapf(err, "failed to open directory %v", prefixStr)})
			return
		}

		log.Log().Strs("dir", dir.GetPath()).Msg("sending directory")
		if !r.SendDir(out, DirErr{Dir: dir}) {
			return
		}
	}
}

func (r *Stream) goReadRange(tr facade.ReadTransaction, query q.Tuple, opts RangeOpts, in chan DirErr, out chan DirKVErr) {
	log := r.Log.With().Str("stage", "read range").Interface("query", query).Logger()

	prefix := toTuplePrefix(query)
	prefix = removeMaybeMore(prefix)
	fdbPrefix, err := convert.ToFDBTuple(prefix)
	if err != nil {
		r.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to convert prefix to FDB tuple")})
		return
	}

	for msg := range in {
		if msg.Err != nil {
			r.SendDirKV(out, DirKVErr{Err: errors.Wrap(msg.Err, "read range input closed")})
			return
		}

		dir := msg.Dir
		log := log.With().Strs("dir", dir.GetPath()).Logger()
		log.Log().Msg("received directory")

		rng, err := fdb.PrefixRange(dir.Pack(fdbPrefix))
		if err != nil {
			r.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to create prefix range")})
			return
		}

		iter := tr.GetRange(rng, fdb.RangeOptions{
			Reverse: opts.Reverse,
			Limit:   opts.Limit,
		}).Iterator()

		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				r.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to get key-value")})
				return
			}
			if !r.SendDirKV(out, DirKVErr{Dir: dir, KV: kv}) {
				return
			}
		}
	}
}

func (r *Stream) goUnpackKeys(query q.Tuple, filter bool, in chan DirKVErr, out chan KeyValErr) {
	log := r.Log.With().Str("stage", "unpack keys").Interface("query", query).Logger()

	for msg := range in {
		if msg.Err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(msg.Err, "filter keys input closed")})
			return
		}

		dir := msg.Dir
		fromDB := msg.KV
		log := log.With().Interface("dir", dir.GetPath()).Logger()
		log.Log().Msg("received key-value")

		tup, err := dir.Unpack(fromDB.Key)
		if err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to unpack key")})
			return
		}

		kv := q.KeyValue{
			Key: q.Key{
				Directory: convert.FromStringArray(dir.GetPath()),
				Tuple:     convert.FromFDBTuple(tup),
			},
			Value: q.Bytes(fromDB.Value),
		}

		if mismatch := compare.Tuples(query, kv.Key.Tuple); mismatch != nil {
			if filter {
				continue
			}
			r.SendKV(out, KeyValErr{Err: errors.Errorf("key's tuple disobeys schema at index path %v", mismatch)})
			return
		}

		log.Log().Msg("sending key-value")
		if !r.SendKV(out, KeyValErr{KV: kv}) {
			return
		}
	}
}

func (r *Stream) goUnpackValues(query q.Value, valHandler internal.ValHandler, in chan KeyValErr, out chan KeyValErr) {
	log := r.Log.With().Str("stage", "unpack values").Interface("query", query).Logger()

	for msg := range in {
		if msg.Err != nil {
			r.SendKV(out, KeyValErr{Err: msg.Err})
			return
		}

		kv := msg.KV
		log := log.With().Interface("kv", kv).Logger()
		log.Log().Msg("received key-value")

		var err error
		kv.Value, err = valHandler.Handle(kv.Value.(q.Bytes))
		if err != nil {
			r.SendKV(out, KeyValErr{Err: err})
			return
		}
		if kv.Value != nil {
			log.Log().Interface("kv", kv).Msg("sending key-value")
			if !r.SendKV(out, KeyValErr{KV: kv}) {
				return
			}
		}
	}
}

func splitAtFirstVariable(dir q.Directory) (q.Directory, *q.Variable, q.Directory) {
	for i, element := range dir {
		if variable, ok := element.(q.Variable); ok {
			return dir[:i], &variable, dir[i+1:]
		}
	}
	return dir, nil, nil
}

func toTuplePrefix(tup q.Tuple) q.Tuple {
	for i, element := range tup {
		if _, ok := element.(q.Variable); ok {
			return tup[:i]
		}
	}
	return tup
}

func removeMaybeMore(tup q.Tuple) q.Tuple {
	if len(tup) > 0 {
		last := len(tup) - 1
		if _, ok := tup[last].(q.MaybeMore); ok {
			tup = tup[:last]
		}
	}
	return tup
}
