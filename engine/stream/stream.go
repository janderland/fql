package stream

import (
	"bytes"
	"context"
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Stream struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	log    *zerolog.Logger
	errCh  chan error
}

type KeyValErr struct {
	KV  keyval.KeyValue
	Err error
}

func New(ctx context.Context) Stream {
	ctx, cancel := context.WithCancel(ctx)

	return Stream{
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
		log:    zerolog.Ctx(ctx),
		errCh:  make(chan error),
	}
}

func (r *Stream) WithErrors(kvCh chan keyval.KeyValue) chan KeyValErr {
	outCh := make(chan KeyValErr)

	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	go func() {
		defer close(outCh)
		for {
			select {
			case err, open := <-r.errCh:
				if !open {
					return
				}
				outCh <- KeyValErr{Err: err}

			case kv, open := <-kvCh:
				if !open {
					kvCh = nil
					continue
				}
				outCh <- KeyValErr{KV: kv}
			}
		}
	}()

	return outCh
}

func (r *Stream) OpenDirectories(tr fdb.ReadTransactor, query keyval.KeyValue) chan directory.DirectorySubspace {
	out := make(chan directory.DirectorySubspace)
	r.wg.Add(1)

	go func() {
		defer close(out)
		defer r.wg.Done()
		r.doOpenDirectories(tr, query.Key.Directory, out)
	}()

	return out
}

func (r *Stream) ReadRange(tr fdb.ReadTransaction, query keyval.KeyValue, in chan directory.DirectorySubspace) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	r.wg.Add(1)

	go func() {
		defer close(out)
		defer r.wg.Done()
		r.doReadRange(tr, query.Key.Tuple, in, out)
	}()

	return out
}

func (r *Stream) FilterKeys(query keyval.KeyValue, in chan keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	r.wg.Add(1)

	go func() {
		defer close(out)
		defer r.wg.Done()
		r.doFilterKeys(query.Key.Tuple, in, out)
	}()

	return out
}

func (r *Stream) UnpackValues(query keyval.KeyValue, in chan keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	r.wg.Add(1)

	go func() {
		defer close(out)
		defer r.wg.Done()
		r.doUnpackValues(query.Value, in, out)
	}()

	return out
}

func (r *Stream) doOpenDirectories(tr fdb.ReadTransactor, query keyval.Directory, out chan directory.DirectorySubspace) {
	log := r.log.With().Str("stage", "open directories").Interface("query", query).Logger()

	prefix, variable, suffix := keyval.SplitAtFirstVariable(query)
	prefixStr, err := keyval.ToStringArray(prefix)
	if err != nil {
		r.sendError(errors.Wrapf(err, "failed to convert directory prefix to string array"))
	}

	if variable != nil {
		subDirs, err := directory.List(tr, prefixStr)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to list directories"))
			return
		}
		if len(subDirs) == 0 {
			r.sendError(errors.Errorf("no subdirectories for %v", prefixStr))
		}

		log.Trace().Strs("sub dirs", subDirs).Msg("found subdirectories")

		for _, subDir := range subDirs {
			var dir keyval.Directory
			dir = append(dir, prefix...)
			dir = append(dir, subDir)
			dir = append(dir, suffix...)
			r.doOpenDirectories(tr, dir, out)
		}
	} else {
		dir, err := directory.Open(tr, prefixStr, nil)
		if err != nil {
			r.sendError(errors.Wrapf(err, "failed to open directory %v", prefixStr))
			return
		}

		log.Debug().Strs("dir", dir.GetPath()).Msg("sending directory")
		r.sendDir(out, dir)
	}
}

func (r *Stream) doReadRange(tr fdb.ReadTransaction, query keyval.Tuple, in chan directory.DirectorySubspace, out chan keyval.KeyValue) {
	log := r.log.With().Str("stage", "read range").Interface("query", query).Logger()

	prefix, _, _ := keyval.SplitAtFirstVariable(query)
	fdbPrefix := keyval.ToFDBTuple(prefix)

	for dir := r.recvDir(in); dir != nil; dir = r.recvDir(in) {
		log := log.With().Strs("dir", dir.GetPath()).Logger()
		log.Debug().Msg("received directory")

		rng, err := fdb.PrefixRange(dir.Pack(fdbPrefix))
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to create prefix range"))
			return
		}

		iter := tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			fromDB, err := iter.Get()
			if err != nil {
				r.sendError(errors.Wrap(err, "failed to get key-value"))
				return
			}

			tup, err := dir.Unpack(fromDB.Key)
			if err != nil {
				r.sendError(errors.Wrap(err, "failed to unpack key"))
				return
			}

			kv := keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.FromStringArray(dir.GetPath()),
					Tuple:     keyval.FromFDBTuple(tup),
				},
				Value: fromDB.Value,
			}

			log.Debug().Interface("kv", kv).Msg("sending key-value")
			r.sendKV(out, kv)
		}
	}
}

func (r *Stream) doFilterKeys(query keyval.Tuple, in chan keyval.KeyValue, out chan keyval.KeyValue) {
	log := r.log.With().Str("stage", "filter keys").Interface("query", query).Logger()

	for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
		log := log.With().Interface("kv", kv).Logger()
		log.Debug().Msg("received key-value")

		if keyval.CompareTuples(query, kv.Key.Tuple) == nil {
			log.Debug().Msg("sending key-value")
			r.sendKV(out, *kv)
		}
	}
}

func (r *Stream) doUnpackValues(query keyval.Value, in chan keyval.KeyValue, out chan keyval.KeyValue) {
	log := r.log.With().Str("stage", "unpack values").Interface("query", query).Logger()

	if variable, isVar := query.(keyval.Variable); isVar {
		for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
			log := log.With().Interface("kv", kv).Logger()
			log.Debug().Msg("received key-value")

			for _, typ := range variable {
				outVal, err := keyval.UnpackValue(typ, kv.Value.([]byte))
				if err != nil {
					continue
				}

				kv.Value = outVal
				log.Debug().Interface("kv", kv).Msg("sending key-value")
				r.sendKV(out, *kv)
				break
			}
		}
	} else {
		queryBytes, err := keyval.PackValue(query)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to pack query value"))
			return
		}

		for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
			log := log.With().Interface("kv", kv).Logger()
			log.Debug().Msg("received key-value")

			if bytes.Equal(queryBytes, kv.Value.([]byte)) {
				kv.Value = query
				log.Debug().Interface("kv", kv).Msg("sending key-value")
				r.sendKV(out, *kv)
			}
		}
	}
}

func (r *Stream) sendDir(ch chan<- directory.DirectorySubspace, dir directory.DirectorySubspace) {
	select {
	case <-r.ctx.Done():
	case ch <- dir:
	}
}

func (r *Stream) recvDir(ch <-chan directory.DirectorySubspace) directory.DirectorySubspace {
	select {
	case <-r.ctx.Done():
	case dir, open := <-ch:
		if open {
			return dir
		}
	}
	return nil
}

func (r *Stream) sendKV(ch chan<- keyval.KeyValue, kv keyval.KeyValue) {
	select {
	case <-r.ctx.Done():
	case ch <- kv:
	}
}

func (r *Stream) recvKV(ch <-chan keyval.KeyValue) *keyval.KeyValue {
	select {
	case <-r.ctx.Done():
	case kv, open := <-ch:
		if open {
			return &kv
		}
	}
	return nil
}

func (r *Stream) sendError(err error) {
	select {
	case <-r.ctx.Done():
	case r.errCh <- err:
		r.cancel()
	}
}
