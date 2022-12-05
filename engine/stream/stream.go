// Package stream performs range-reads and filtering.
// TODO: Examples.
package stream

import (
	"context"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/engine/internal"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/compare"
	"github.com/janderland/fdbq/keyval/convert"
)

type (
	// RangeOpts specifies how a Stream executes a query.
	RangeOpts struct {
		Reverse bool
		Limit   int
	}

	// Stream provides methods which build pipelines for reading
	// a range of key-values. ctx controls the cancellation of all
	// operations. For the methods which spawn a goroutine, canceling
	// ctx will stop them. For the methods which block on sending to
	// a channel, canceling ctx will unblock them.
	Stream struct {
		ctx context.Context
		log zerolog.Logger
	}

	// DirErr is streamed from a call to [Stream.OpenDirectories].
	// If Err is nil, the other fields should be non-nil. If Err
	// is non-nil, the other fields should be nil.
	DirErr struct {
		Dir directory.DirectorySubspace
		Err error
	}

	// DirKVErr is streamed from a call to [Stream.ReadRange].
	// If Err is nil, the other fields should be non-nil. If
	// Err is non-nil, the other fields should be nil.
	DirKVErr struct {
		Dir directory.DirectorySubspace
		KV  fdb.KeyValue
		Err error
	}

	// KeyValErr is streamed from a call to [Stream.UnpackKeys] or
	// [Stream.UnpackValues]. If Err is nil, the other fields should
	// be non-nil. If Err is non-nil, the other fields should be nil.
	KeyValErr struct {
		KV  keyval.KeyValue
		Err error
	}
)

// New constructs a new Stream. The context provides a way
// to cancel any pipelines created with this Stream.
func New(ctx context.Context, log zerolog.Logger) Stream {
	return Stream{
		ctx: ctx,
		log: log,
	}
}

// SendDir sends the given DirErr onto the given channel and returns
// true. If the context.Context associated with this Stream is canceled,
// then nothing is sent and false is returned.
func (x *Stream) SendDir(out chan<- DirErr, in DirErr) bool {
	select {
	case <-x.ctx.Done():
		return false
	case out <- in:
		return true
	}
}

// SendDirKV sends the given DirKVErr onto the given channel and returns
// true. If the context.Context associated with this Stream is canceled,
// then nothing is sent and false is returned.
func (x *Stream) SendDirKV(out chan<- DirKVErr, in DirKVErr) bool {
	select {
	case <-x.ctx.Done():
		return false
	case out <- in:
		return true
	}
}

// SendKV sends the given KeyValErr onto the given channel and returns
// true. If the context.Context associated with this Stream is canceled,
// then nothing is sent and false is returned.
func (x *Stream) SendKV(out chan<- KeyValErr, in KeyValErr) bool {
	select {
	case <-x.ctx.Done():
		return false
	case out <- in:
		return true
	}
}

// OpenDirectories executes the given directory query in a separate goroutine using the given
// transactor. When the goroutine exits, the returned channel is closed. If the associated
// context.Context is canceled, then the goroutine exits after the latest FDB call.
func (x *Stream) OpenDirectories(tr facade.ReadTransactor, query keyval.Directory) chan DirErr {
	out := make(chan DirErr)

	go func() {
		defer close(out)
		x.goOpenDirectories(tr, query, out)
	}()

	return out
}

// ReadRange executes range-reads in a separate goroutine using the given transactor. When the goroutine exits, the
// returned channel is closed. Any errors read from the input channel are wrapped and forwarded. For each directory
// read from the input channel, a range-read is performed using the tuple prefix defined by the given [keyval.Tuple].
// If the associated context.Context is canceled, then the goroutine exits after the latest FDB call.
func (x *Stream) ReadRange(tr facade.ReadTransaction, query keyval.Tuple, opts RangeOpts, in chan DirErr) chan DirKVErr {
	out := make(chan DirKVErr)

	go func() {
		defer close(out)
		x.goReadRange(tr, query, opts, in, out)
	}()

	return out
}

// UnpackKeys converts the channel of DirKVErr into a channel of KeyValErr in a separate goroutine.
// When the goroutine exits, the returned channel is closed. Any errors read from the input channel
// are wrapped and forwarded. Keys are unpacked using subspace.Subspace.Unpack and then converted to
// FDBQ types. Values are converted to [keyval.Bytes]; the actual byte string remains unchanged.
func (x *Stream) UnpackKeys(query keyval.Tuple, filter bool, in chan DirKVErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		x.goUnpackKeys(query, filter, in, out)
	}()

	return out
}

// UnpackValues deserializes the values in a separate goroutine. When the goroutine exits, the returned channel
// is closed. Any errors read from the input channel are wrapped and forwarded. The values of the key-values
// provided via the input channel are expected to be of type [keyval.Bytes], and are converted to the type specified
// in the given schema.
func (x *Stream) UnpackValues(query keyval.Value, valHandler internal.ValHandler, in chan KeyValErr) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		x.goUnpackValues(query, valHandler, in, out)
	}()

	return out
}

func (x *Stream) goOpenDirectories(tr facade.ReadTransactor, query keyval.Directory, out chan DirErr) {
	log := x.log.With().Str("stage", "open directories").Interface("query", query).Logger()

	prefix, variable, suffix := splitAtFirstVariable(query)
	prefixStr, err := convert.ToStringArray(prefix)
	if err != nil {
		x.SendDir(out, DirErr{Err: errors.Wrapf(err, "failed to convert directory prefix to string array")})
		return
	}

	if variable != nil {
		subDirs, err := tr.DirList(prefixStr)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return
			}
			x.SendDir(out, DirErr{Err: errors.Wrap(err, "failed to list directories")})
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
			if err := x.ctx.Err(); err != nil {
				x.SendDir(out, DirErr{Err: err})
				return
			}

			var dir keyval.Directory
			dir = append(dir, prefix...)
			dir = append(dir, keyval.String(subDir))
			dir = append(dir, suffix...)
			x.goOpenDirectories(tr, dir, out)
		}
	} else {
		dir, err := tr.DirOpen(prefixStr)
		if err != nil {
			if errors.Is(err, directory.ErrDirNotExists) {
				return
			}
			x.SendDir(out, DirErr{Err: errors.Wrapf(err, "failed to open directory %v", prefixStr)})
			return
		}

		log.Log().Strs("dir", dir.GetPath()).Msg("sending directory")
		if !x.SendDir(out, DirErr{Dir: dir}) {
			return
		}
	}
}

func (x *Stream) goReadRange(tr facade.ReadTransaction, query keyval.Tuple, opts RangeOpts, in chan DirErr, out chan DirKVErr) {
	log := x.log.With().Str("stage", "read range").Interface("query", query).Logger()

	prefix := toTuplePrefix(query)
	prefix = removeMaybeMore(prefix)
	fdbPrefix, err := convert.ToFDBTuple(prefix)
	if err != nil {
		x.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to convert prefix to FDB tuple")})
		return
	}

	for msg := range in {
		if msg.Err != nil {
			x.SendDirKV(out, DirKVErr{Err: errors.Wrap(msg.Err, "read range input closed")})
			return
		}

		dir := msg.Dir
		log := log.With().Strs("dir", dir.GetPath()).Logger()
		log.Log().Msg("received directory")

		rng, err := fdb.PrefixRange(dir.Pack(fdbPrefix))
		if err != nil {
			x.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to create prefix range")})
			return
		}

		iter := tr.GetRange(rng, fdb.RangeOptions{
			Reverse: opts.Reverse,
			Limit:   opts.Limit,
		}).Iterator()

		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				x.SendDirKV(out, DirKVErr{Err: errors.Wrap(err, "failed to get key-value")})
				return
			}
			if !x.SendDirKV(out, DirKVErr{Dir: dir, KV: kv}) {
				return
			}
		}
	}
}

func (x *Stream) goUnpackKeys(query keyval.Tuple, filter bool, in chan DirKVErr, out chan KeyValErr) {
	log := x.log.With().Str("stage", "unpack keys").Interface("query", query).Logger()

	for msg := range in {
		if msg.Err != nil {
			x.SendKV(out, KeyValErr{Err: errors.Wrap(msg.Err, "filter keys input closed")})
			return
		}

		dir := msg.Dir
		fromDB := msg.KV
		log := log.With().Interface("dir", dir.GetPath()).Logger()
		log.Log().Msg("received key-value")

		tup, err := dir.Unpack(fromDB.Key)
		if err != nil {
			x.SendKV(out, KeyValErr{Err: errors.Wrap(err, "failed to unpack key")})
			return
		}

		kv := keyval.KeyValue{
			Key: keyval.Key{
				Directory: convert.FromStringArray(dir.GetPath()),
				Tuple:     convert.FromFDBTuple(tup),
			},
			Value: keyval.Bytes(fromDB.Value),
		}

		if mismatch := compare.Tuples(query, kv.Key.Tuple); mismatch != nil {
			if filter {
				continue
			}
			x.SendKV(out, KeyValErr{Err: errors.Errorf("key's tuple disobeys schema at index path %v", mismatch)})
			return
		}

		log.Log().Msg("sending key-value")
		if !x.SendKV(out, KeyValErr{KV: kv}) {
			return
		}
	}
}

func (x *Stream) goUnpackValues(query keyval.Value, valHandler internal.ValHandler, in chan KeyValErr, out chan KeyValErr) {
	log := x.log.With().Str("stage", "unpack values").Interface("query", query).Logger()

	for msg := range in {
		if msg.Err != nil {
			x.SendKV(out, KeyValErr{Err: msg.Err})
			return
		}

		kv := msg.KV
		log := log.With().Interface("kv", kv).Logger()
		log.Log().Msg("received key-value")

		var err error
		kv.Value, err = valHandler.Handle(kv.Value.(keyval.Bytes))
		if err != nil {
			x.SendKV(out, KeyValErr{Err: err})
			return
		}
		if kv.Value != nil {
			log.Log().Interface("kv", kv).Msg("sending key-value")
			if !x.SendKV(out, KeyValErr{KV: kv}) {
				return
			}
		}
	}
}

func splitAtFirstVariable(dir keyval.Directory) (keyval.Directory, *keyval.Variable, keyval.Directory) {
	for i, element := range dir {
		if variable, ok := element.(keyval.Variable); ok {
			return dir[:i], &variable, dir[i+1:]
		}
	}
	return dir, nil, nil
}

func toTuplePrefix(tup keyval.Tuple) keyval.Tuple {
	for i, element := range tup {
		if _, ok := element.(keyval.Variable); ok {
			return tup[:i]
		}
	}
	return tup
}

func removeMaybeMore(tup keyval.Tuple) keyval.Tuple {
	if len(tup) > 0 {
		last := len(tup) - 1
		if _, ok := tup[last].(keyval.MaybeMore); ok {
			tup = tup[:last]
		}
	}
	return tup
}
