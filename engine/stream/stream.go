package stream

import (
	"bytes"
	"context"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type (
	Stream struct {
		ctx context.Context
		log *zerolog.Logger
	}

	DirErr struct {
		Dir directory.DirectorySubspace
		Err error
	}

	KeyValErr struct {
		KV  keyval.KeyValue
		Err error
	}
)

func New(ctx context.Context) (Stream, func()) {
	ctx, cancel := context.WithCancel(ctx)

	return Stream{
		ctx: ctx,
		log: zerolog.Ctx(ctx),
	}, cancel
}

func (r *Stream) SendDir(out chan<- DirErr, in DirErr) bool {
	select {
	case <-r.ctx.Done():
		return false
	case out <- in:
		return true
	}
}

func (r *Stream) SendKV(out chan<- KeyValErr, in KeyValErr) bool {
	select {
	case <-r.ctx.Done():
		return false
	case out <- in:
		return true
	}
}

func (r *Stream) OpenDirectories(tr fdb.ReadTransactor, query keyval.KeyValue) chan DirErr {
	out := make(chan DirErr)

	go func() {
		defer close(out)
		r.doOpenDirectories(tr, query.Key.Directory, out)
	}()

	return out
}

func (r *Stream) ReadRange(tr fdb.ReadTransaction, query keyval.KeyValue, in chan DirErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		r.doReadRange(tr, query.Key.Tuple, in, out)
	}()

	return out
}

func (r *Stream) FilterKeys(query keyval.KeyValue, in chan KeyValErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		r.doFilterKeys(query.Key.Tuple, in, out)
	}()

	return out
}

func (r *Stream) UnpackValues(query keyval.KeyValue, in chan KeyValErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		r.doUnpackValues(query.Value, in, out)
	}()

	return out
}

func (r *Stream) doOpenDirectories(tr fdb.ReadTransactor, query keyval.Directory, out chan DirErr) {
	log := r.log.With().Str("stage", "open directories").Interface("query", query).Logger()

	prefix, variable, suffix := keyval.SplitAtFirstVariable(query)
	prefixStr, err := keyval.ToStringArray(prefix)
	if err != nil {
		r.SendDir(out, DirErr{Err: errors.Wrapf(err, "failed to convert directory prefix to string array")})
		return
	}

	if variable != nil {
		subDirs, err := directory.List(tr, prefixStr)
		if err != nil {
			r.SendDir(out, DirErr{Err: errors.Wrap(err, "failed to list directories")})
			return
		}
		if len(subDirs) == 0 {
			r.SendDir(out, DirErr{Err: errors.Errorf("no subdirectories for %v", prefixStr)})
			return
		}

		log.Trace().Strs("sub dirs", subDirs).Msg("found subdirectories")

		for _, subDir := range subDirs {
			// Between each interaction with the DB, give
			// this goroutine a chance to exit early.
			if err := r.ctx.Err(); err != nil {
				r.SendDir(out, DirErr{Err: err})
				return
			}

			var dir keyval.Directory
			dir = append(dir, prefix...)
			dir = append(dir, subDir)
			dir = append(dir, suffix...)
			r.doOpenDirectories(tr, dir, out)
		}
	} else {
		dir, err := directory.Open(tr, prefixStr, nil)
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

func (r *Stream) doReadRange(tr fdb.ReadTransaction, query keyval.Tuple, in chan DirErr, out chan KeyValErr) {
	log := r.log.With().Str("stage", "read range").Interface("query", query).Logger()

	prefix, _, _ := keyval.SplitAtFirstVariable(query)
	fdbPrefix := keyval.ToFDBTuple(prefix)

	for msg := range in {
		if msg.Err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(msg.Err, "read range input closed")})
			return
		}

		dir := msg.Dir
		log := log.With().Strs("dir", dir.GetPath()).Logger()
		log.Log().Msg("received directory")

		rng, err := fdb.PrefixRange(dir.Pack(fdbPrefix))
		if err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to create prefix range")})
			return
		}

		iter := tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			fromDB, err := iter.Get()
			if err != nil {
				r.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to get key-value")})
				return
			}

			tup, err := dir.Unpack(fromDB.Key)
			if err != nil {
				r.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to unpack key")})
				return
			}

			kv := keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.FromStringArray(dir.GetPath()),
					Tuple:     keyval.FromFDBTuple(tup),
				},
				Value: fromDB.Value,
			}

			log.Log().Interface("kv", kv).Msg("sending key-value")
			if !r.SendKV(out, KeyValErr{KV: kv}) {
				return
			}
		}
	}
}

func (r *Stream) doFilterKeys(query keyval.Tuple, in chan KeyValErr, out chan KeyValErr) {
	log := r.log.With().Str("stage", "filter keys").Interface("query", query).Logger()

	for msg := range in {
		if msg.Err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(msg.Err, "filter keys input closed")})
			return
		}

		kv := msg.KV
		log := log.With().Interface("kv", kv).Logger()
		log.Log().Msg("received key-value")

		if keyval.CompareTuples(query, kv.Key.Tuple) == nil {
			log.Log().Msg("sending key-value")
			if !r.SendKV(out, KeyValErr{KV: kv}) {
				return
			}
		}
	}
}

func (r *Stream) doUnpackValues(query keyval.Value, in chan KeyValErr, out chan KeyValErr) {
	log := r.log.With().Str("stage", "unpack values").Interface("query", query).Logger()

	if variable, isVar := query.(keyval.Variable); isVar {
		for msg := range in {
			if msg.Err != nil {
				r.SendKV(out, KeyValErr{Err: msg.Err})
				return
			}

			kv := msg.KV
			log := log.With().Interface("kv", kv).Logger()
			log.Log().Msg("received key-value")

			if len(variable) == 0 {
				if !r.SendKV(out, KeyValErr{KV: kv}) {
					return
				}
				continue
			}

			for _, typ := range variable {
				outVal, err := keyval.UnpackValue(typ, kv.Value.([]byte))
				if err != nil {
					continue
				}

				kv.Value = outVal
				log.Log().Interface("kv", kv).Msg("sending key-value")
				if !r.SendKV(out, KeyValErr{KV: kv}) {
					return
				}
				break
			}
		}
	} else {
		queryBytes, err := keyval.PackValue(query)
		if err != nil {
			r.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to pack query value")})
			return
		}

		for msg := range in {
			if msg.Err != nil {
				r.SendKV(out, KeyValErr{Err: msg.Err})
				return
			}

			kv := msg.KV
			log := log.With().Interface("kv", kv).Logger()
			log.Log().Msg("received key-value")

			if bytes.Equal(queryBytes, kv.Value.([]byte)) {
				kv.Value = query
				log.Log().Interface("kv", kv).Msg("sending key-value")
				if !r.SendKV(out, KeyValErr{KV: kv}) {
					return
				}
			}
		}
	}
}
