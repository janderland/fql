package reader

import (
	"context"
	"math/big"
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	dir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
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

type DirKeyValue struct {
	dir dir.DirectorySubspace
	kv  fdb.KeyValue
}

type DirTupValue struct {
	dir dir.DirectorySubspace
	tup tup.Tuple
	val []byte
}

func New(tr fdb.Transaction) Reader {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	return Reader{
		tr:     tr,
		wg:     &wg,
		ctx:    ctx,
		cancel: cancel,
		errCh:  make(chan error),
	}
}

func (r *Reader) Read(kv keyval.KeyValue) (chan DirTupValue, chan error) {
	dirCh := r.openDirectories(kv.Key.Directory)
	dkvCh := r.readRange(kv.Key.Tuple, dirCh)
	dtvCh := r.filterRange(kv.Key.Tuple, dkvCh)

	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	return dtvCh, r.errCh
}

func (r *Reader) signalError(err error) {
	select {
	case r.errCh <- err:
		r.cancel()
	case <-r.ctx.Done():
	}
}

func (r *Reader) openDirectories(directory keyval.Directory) chan dir.DirectorySubspace {
	dirCh := make(chan dir.DirectorySubspace)
	r.wg.Add(1)

	go func() {
		defer close(dirCh)
		defer r.wg.Done()
		r.doOpenDirectories(directory, dirCh)
	}()

	return dirCh
}

func (r *Reader) doOpenDirectories(directory keyval.Directory, dirCh chan dir.DirectorySubspace) {
	prefix, variable, suffix := splitAtFirstVariable(directory)
	prefixStr := toStringArray(prefix)

	if variable != nil {
		subDirs, err := dir.List(r.tr, prefixStr)
		if err != nil {
			r.signalError(errors.Wrap(err, "failed to list directories"))
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
			r.signalError(errors.Wrap(err, "failed to open directory"))
			return
		}

		select {
		case <-r.ctx.Done():
			return
		case dirCh <- directory:
		}
	}
}

func (r *Reader) readRange(tuple keyval.Tuple, dirCh chan dir.DirectorySubspace) chan DirKeyValue {
	kvCh := make(chan DirKeyValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doReadRange(tuple, dirCh, kvCh)
		}()
	}

	go func() {
		defer close(kvCh)
		wg.Wait()
	}()

	return kvCh
}

func (r *Reader) doReadRange(tuple keyval.Tuple, dirCh chan dir.DirectorySubspace, kvCh chan DirKeyValue) {
	read := func() (dir.DirectorySubspace, bool) {
		select {
		case <-r.ctx.Done():
			return nil, false
		case directory, open := <-dirCh:
			return directory, open
		}
	}

	prefix, _, _ := splitAtFirstVariable(tuple)
	fdbPrefix := toFDBTuple(prefix)

	for directory, running := read(); running; directory, running = read() {
		rng, err := fdb.PrefixRange(directory.Pack(fdbPrefix))
		if err != nil {
			r.signalError(errors.Wrap(err, "failed to create prefix range"))
			return
		}

		iter := r.tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				r.signalError(errors.Wrap(err, "failed to get key-value"))
				return
			}

			select {
			case <-r.ctx.Done():
				return
			case kvCh <- DirKeyValue{dir: directory, kv: kv}:
			}
		}
	}
}

func (r *Reader) filterRange(tuple keyval.Tuple, in chan DirKeyValue) chan DirTupValue {
	out := make(chan DirTupValue)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		r.wg.Add(1)
		wg.Add(1)

		go func() {
			defer r.wg.Done()
			defer wg.Done()
			r.doFilterRange(tuple, in, out)
		}()
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func (r *Reader) doFilterRange(pattern keyval.Tuple, in chan DirKeyValue, out chan DirTupValue) {
	read := func() (DirKeyValue, bool) {
		select {
		case <-r.ctx.Done():
			return DirKeyValue{}, false
		case directory, open := <-in:
			return directory, open
		}
	}

	for dkv, running := read(); running; dkv, running = read() {
		tuple, err := dkv.dir.Unpack(dkv.kv.Key)
		if err != nil {
			r.signalError(errors.Wrap(err, "failed to unpack key"))
		}

		if compareTuples(pattern, tuple) == -1 {
			select {
			case <-r.ctx.Done():
				return
			case out <- DirTupValue{dir: dkv.dir, tup: tuple, val: dkv.kv.Value}:
			}
		}
	}
}

func compareTuples(pattern keyval.Tuple, candidate tup.Tuple) int {
	if len(pattern) < len(candidate) {
		return len(pattern)
	}
	if len(pattern) > len(candidate) {
		return len(candidate)
	}

	var i int
	var e tup.TupleElement
	err := keyval.ParseTuple(candidate, func(p *keyval.TupleParser) {
		for i, e = range pattern {
			switch e.(type) {
			// TODO: Ensure all possible FDB tuple elements are covered.
			case int64:
				if p.Int() != e.(int64) {
					return
				}
			case uint64:
				if p.Uint() != e.(uint64) {
					return
				}
			case string:
				if p.String() != e.(string) {
					return
				}
			case *big.Int:
				if p.BigInt().Cmp(e.(*big.Int)) != 0 {
					return
				}
			case float32:
				if p.Float() != float64(e.(float32)) {
					return
				}
			case float64:
				if p.Float() != e.(float64) {
					return
				}
			case bool:
				if p.Bool() != e.(bool) {
					return
				}
			}
		}
	})
	if err != nil {
		return i
	}
	return -1
}

func splitAtFirstVariable(list []interface{}) ([]interface{}, *keyval.Variable, []interface{}) {
	for i, segment := range list {
		switch segment.(type) {
		case keyval.Variable:
			v := segment.(keyval.Variable)
			return list[:i], &v, list[i+1:]
		}
	}
	return list, nil, nil
}

func toStringArray(in []interface{}) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = in[i].(string)
	}
	return out
}

func toFDBTuple(in []interface{}) tup.Tuple {
	out := make(tup.Tuple, len(in))
	for i := range in {
		out[i] = tup.TupleElement(in[i])
	}
	return out
}
