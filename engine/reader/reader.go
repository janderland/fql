package reader

import (
	"bytes"
	"context"
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type Reader struct {
	tr     fdb.Transaction
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	errCh  chan error
}

func New(tr fdb.Transaction) Reader {
	ctx, cancel := context.WithCancel(context.Background())

	return Reader{
		tr:     tr,
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
		errCh:  make(chan error),
	}
}

func (r *Reader) Read(query keyval.KeyValue) (chan keyval.KeyValue, chan error) {
	stage1 := r.openDirectories(query)
	stage2 := r.readRange(query, stage1)
	stage3 := r.filterKeys(query, stage2)
	out := r.unpackValues(query, stage3)

	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	return out, r.errCh
}

func (r *Reader) sendDir(ch chan<- directory.DirectorySubspace, dir directory.DirectorySubspace) {
	select {
	case <-r.ctx.Done():
	case ch <- dir:
	}
}

func (r *Reader) recvDir(ch <-chan directory.DirectorySubspace) directory.DirectorySubspace {
	select {
	case <-r.ctx.Done():
	case dir, open := <-ch:
		if open {
			return dir
		}
	}
	return nil
}

func (r *Reader) sendKV(ch chan<- keyval.KeyValue, kv keyval.KeyValue) {
	select {
	case <-r.ctx.Done():
	case ch <- kv:
	}
}

func (r *Reader) recvKV(ch <-chan keyval.KeyValue) *keyval.KeyValue {
	select {
	case <-r.ctx.Done():
	case kv, open := <-ch:
		if open {
			return &kv
		}
	}
	return nil
}

func (r *Reader) sendError(err error) {
	select {
	case <-r.ctx.Done():
	case r.errCh <- err:
		r.cancel()
	}
}

func (r *Reader) openDirectories(query keyval.KeyValue) chan directory.DirectorySubspace {
	dirCh := make(chan directory.DirectorySubspace)
	r.wg.Add(1)

	go func() {
		defer close(dirCh)
		defer r.wg.Done()
		r.doOpenDirectories(query.Key.Directory, dirCh)
	}()

	return dirCh
}

func (r *Reader) readRange(query keyval.KeyValue, in chan directory.DirectorySubspace) chan keyval.KeyValue {
	kvCh := make(chan keyval.KeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doReadRange(query.Key.Tuple, in, kvCh)
		}()
	}

	go func() {
		defer close(kvCh)
		wg.Wait()
	}()

	return kvCh
}

func (r *Reader) filterKeys(query keyval.KeyValue, in chan keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doFilterKeys(query.Key.Tuple, in, out)
		}()
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func (r *Reader) unpackValues(query keyval.KeyValue, in chan keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doUnpackValues(query.Value, in, out)
		}()
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func (r *Reader) doOpenDirectories(query keyval.Directory, out chan directory.DirectorySubspace) {
	prefix, variable, suffix := keyval.SplitAtFirstVariable(query)
	prefixStr, err := keyval.ToStringArray(prefix)
	if err != nil {
		r.sendError(errors.Wrapf(err, "fail to convert directory prefix to string array"))
	}

	if variable != nil {
		subDirs, err := directory.List(r.tr, prefixStr)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to list directories"))
			return
		}

		for _, subDir := range subDirs {
			var dir keyval.Directory
			dir = append(dir, prefix...)
			dir = append(dir, subDir)
			dir = append(dir, suffix...)
			r.doOpenDirectories(dir, out)
		}
	} else {
		dir, err := directory.Open(r.tr, prefixStr, nil)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to open directory"))
			return
		}
		r.sendDir(out, dir)
	}
}

func (r *Reader) doReadRange(query keyval.Tuple, in chan directory.DirectorySubspace, out chan keyval.KeyValue) {
	prefix, _, _ := keyval.SplitAtFirstVariable(query)
	fdbPrefix := keyval.ToFDBTuple(prefix)

	for dir := r.recvDir(in); dir != nil; dir = r.recvDir(in) {
		rng, err := fdb.PrefixRange(dir.Pack(fdbPrefix))
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to create prefix range"))
			return
		}

		iter := r.tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				r.sendError(errors.Wrap(err, "failed to get key-value"))
				return
			}

			tup, err := dir.Unpack(kv.Key)
			if err != nil {
				r.sendError(errors.Wrap(err, "failed to unpack key"))
				return
			}

			r.sendKV(out, keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.FromStringArray(dir.GetPath()),
					Tuple:     keyval.FromFDBTuple(tup),
				},
				Value: kv.Value,
			})
		}
	}
}

func (r *Reader) doFilterKeys(query keyval.Tuple, in chan keyval.KeyValue, out chan keyval.KeyValue) {
	for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
		if keyval.CompareTuples(query, kv.Key.Tuple) == nil {
			r.sendKV(out, *kv)
		}
	}
}

func (r *Reader) doUnpackValues(query keyval.Value, in chan keyval.KeyValue, out chan keyval.KeyValue) {
	if variable, isVar := query.(keyval.Variable); isVar {
		for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
			for _, typ := range variable.Type {
				outVal, err := keyval.UnpackValue(typ, kv.Value.([]byte))
				if err != nil {
					continue
				}
				kv.Value = outVal
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
			if bytes.Compare(queryBytes, kv.Value.([]byte)) == 0 {
				r.sendKV(out, *kv)
			}
		}
	}
}
