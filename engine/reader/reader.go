package reader

import (
	"context"
	"math/big"
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	dir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
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
	out := r.filterValues(query, stage3)

	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	return out, r.errCh
}

func (r *Reader) sendDir(ch chan<- dir.DirectorySubspace, d dir.DirectorySubspace) {
	select {
	case <-r.ctx.Done():
	case ch <- d:
	}
}

func (r *Reader) recvDir(ch <-chan dir.DirectorySubspace) dir.DirectorySubspace {
	select {
	case <-r.ctx.Done():
	case directory, open := <-ch:
		if open {
			return directory
		}
	}
	return nil
}

func (r *Reader) sendKV(ch chan<- keyval.KeyValue, dtv keyval.KeyValue) {
	select {
	case <-r.ctx.Done():
	case ch <- dtv:
	}
}

func (r *Reader) recvKV(ch <-chan keyval.KeyValue) *keyval.KeyValue {
	select {
	case <-r.ctx.Done():
	case dtv, open := <-ch:
		if open {
			return &dtv
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

func (r *Reader) openDirectories(query keyval.KeyValue) chan dir.DirectorySubspace {
	dirCh := make(chan dir.DirectorySubspace)
	r.wg.Add(1)

	go func() {
		defer close(dirCh)
		defer r.wg.Done()
		r.doOpenDirectories(query.Key.Directory, dirCh)
	}()

	return dirCh
}

func (r *Reader) doOpenDirectories(query keyval.Directory, dirCh chan dir.DirectorySubspace) {
	prefix, variable, suffix := keyval.SplitAtFirstVariable(query)
	prefixStr, err := keyval.ToStringArray(prefix)
	if err != nil {
		r.sendError(errors.Wrapf(err, "fail to convert directory prefix to string array"))
	}

	if variable != nil {
		subDirs, err := dir.List(r.tr, prefixStr)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to list directories"))
			return
		}

		for _, sDir := range subDirs {
			var directory keyval.Directory
			directory = append(directory, prefix...)
			directory = append(directory, sDir)
			directory = append(directory, suffix...)
			r.doOpenDirectories(directory, dirCh)
		}
	} else {
		directory, err := dir.Open(r.tr, prefixStr, nil)
		if err != nil {
			r.sendError(errors.Wrap(err, "failed to open directory"))
			return
		}
		r.sendDir(dirCh, directory)
	}
}

func (r *Reader) readRange(query keyval.KeyValue, dirCh chan dir.DirectorySubspace) chan keyval.KeyValue {
	kvCh := make(chan keyval.KeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doReadRange(query.Key.Tuple, dirCh, kvCh)
		}()
	}

	go func() {
		defer close(kvCh)
		wg.Wait()
	}()

	return kvCh
}

func (r *Reader) doReadRange(query keyval.Tuple, dirCh chan dir.DirectorySubspace, kvCh chan keyval.KeyValue) {
	prefix, _, _ := keyval.SplitAtFirstVariable(query)
	fdbPrefix := keyval.ToFDBTuple(prefix)

	for directory := r.recvDir(dirCh); directory != nil; directory = r.recvDir(dirCh) {
		rng, err := fdb.PrefixRange(directory.Pack(fdbPrefix))
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

			tuple, err := directory.Unpack(kv.Key)
			if err != nil {
				r.sendError(errors.Wrap(err, "failed to unpack key"))
				return
			}

			r.sendKV(kvCh, keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.FromStringArray(directory.GetPath()),
					Tuple:     keyval.FromFDBTuple(tuple),
				},
				Value: kv.Value,
			})
		}
	}
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

func (r *Reader) doFilterKeys(query keyval.Tuple, in chan keyval.KeyValue, out chan keyval.KeyValue) {
	for kv := r.recvKV(in); kv != nil; kv = r.recvKV(in) {
		mismatch := compareTuples(query, kv.Key.Tuple)
		if len(mismatch) == 0 {
			r.sendKV(out, *kv)
		}
	}
}

func (r *Reader) filterValues(query keyval.KeyValue, in chan keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doFilterValues(query.Value, in, out)
		}()
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func (r *Reader) doFilterValues(_ keyval.Value, in <-chan keyval.KeyValue, out chan<- keyval.KeyValue) {
	// TODO: Implement value filtering.
	for kv := range in {
		r.sendKV(out, kv)
	}
}

func compareTuples(pattern keyval.Tuple, candidate keyval.Tuple) []int {
	var index []int
	err := keyval.ParseTuple(candidate, func(p *keyval.TupleParser) error {
		for i, e := range pattern {
			switch e.(type) {
			case keyval.Variable:
				// TODO: Check variable constraints.
				break

			case keyval.Tuple:
				if subIndex := compareTuples(e.(keyval.Tuple), p.Tuple()); len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}

			case int64:
				if p.Int() != e.(int64) {
					index = []int{i}
					return nil
				}

			case uint64:
				if p.Uint() != e.(uint64) {
					index = []int{i}
					return nil
				}

			case string:
				if p.String() != e.(string) {
					index = []int{i}
					return nil
				}

			case *big.Int:
				if p.BigInt().Cmp(e.(*big.Int)) != 0 {
					index = []int{i}
					return nil
				}

			case float32:
				if p.Float() != float64(e.(float32)) {
					index = []int{i}
					return nil
				}

			case float64:
				if p.Float() != e.(float64) {
					index = []int{i}
					return nil
				}

			case bool:
				if p.Bool() != e.(bool) {
					index = []int{i}
					return nil
				}

			default:
				if e != p.Any() {
					index = []int{i}
					return nil
				}
			}
		}
		return nil
	})
	if len(index) > 0 {
		return index
	}
	if err != nil {
		if c, ok := err.(keyval.ConversionError); ok {
			return []int{c.Index}
		}
		if err == keyval.ShortTupleError {
			return []int{len(candidate) + 1}
		}
		if err == keyval.LongTupleError {
			return []int{len(pattern) + 1}
		}
	}
	return nil
}
