package engine

import (
	"context"
	"sync"

	dir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

	"github.com/janderland/fdbq/query"
	"github.com/pkg/errors"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

type Coordinator struct {
	tr     fdb.Transaction
	ctx    context.Context
	cancel context.CancelFunc
	errCh  chan error
}

func NewCoordinator(tr fdb.Transaction) Coordinator {
	ctx, cancel := context.WithCancel(context.Background())
	return Coordinator{
		tr:     tr,
		ctx:    ctx,
		cancel: cancel,
		errCh:  make(chan error),
	}
}

func (c *Coordinator) signalError(err error) {
	select {
	case <-c.ctx.Done():
	case c.errCh <- err:
	}
}

func (c *Coordinator) OpenDirectories(directory query.Directory) chan dir.DirectorySubspace {
	dirCh := make(chan dir.DirectorySubspace)

	go func() {
		defer close(dirCh)
		c.openDirectories(directory, dirCh)
	}()

	return dirCh
}

func (c *Coordinator) openDirectories(directory query.Directory, dirCh chan dir.DirectorySubspace) {
	prefix, variable, suffix := splitAtFirstVariable(directory)
	prefixStr := toStringArray(prefix)

	if variable != nil {
		subDirs, err := dir.List(c.tr, prefixStr)
		if err != nil {
			c.signalError(errors.Wrap(err, "failed to list directories"))
			return
		}

		for _, sDir := range subDirs {
			var directory query.Directory
			directory = append(directory, prefix...)
			directory = append(directory, sDir)
			directory = append(directory, suffix...)
			c.openDirectories(directory, dirCh)
		}
	} else {
		directory, err := dir.CreateOrOpen(c.tr, prefixStr, nil)
		if err != nil {
			c.signalError(errors.Wrap(err, "failed to open directory"))
			return
		}

		select {
		case <-c.ctx.Done():
			return
		case dirCh <- directory:
		}
	}
}

type DirKeyValue struct {
	dir dir.DirectorySubspace
	kv  fdb.KeyValue
}

func (c *Coordinator) ReadRange(tuple query.Tuple, dirCh chan dir.DirectorySubspace) chan DirKeyValue {
	kvCh := make(chan DirKeyValue)
	fdbTuple := toFDBTuple(tuple)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.readRange(fdbTuple, dirCh, kvCh)
		}()
	}

	go func() {
		defer close(kvCh)
		wg.Wait()
	}()

	return kvCh
}

func (c *Coordinator) readRange(tuple tup.Tuple, dirCh chan dir.DirectorySubspace, kvCh chan DirKeyValue) {
	read := func() (dir.DirectorySubspace, bool) {
		select {
		case <-c.ctx.Done():
			return nil, false
		case directory, open := <-dirCh:
			return directory, open
		}
	}

	for directory, running := read(); running; directory, running = read() {
		rng, err := fdb.PrefixRange(directory.Pack(tuple))
		if err != nil {
			c.signalError(errors.Wrap(err, "failed to create prefix range"))
			return
		}

		iter := c.tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				c.signalError(errors.Wrap(err, "failed to get key-value"))
				return
			}

			select {
			case <-c.ctx.Done():
				return
			case kvCh <- DirKeyValue{dir: directory, kv: kv}:
			}
		}
	}
}

func (c *Coordinator) FilterRange(tuple query.Tuple, in chan DirKeyValue) chan DirKeyValue {
	out := make(chan DirKeyValue)
	fdbTuple := toFDBTuple(tuple)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.filterRange(fdbTuple, in, out)
		}()
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func (c *Coordinator) filterRange(tuple tup.Tuple, in chan DirKeyValue, out chan DirKeyValue) {
	// TODO
}

func splitAtFirstVariable(list []interface{}) ([]interface{}, *query.Variable, []interface{}) {
	for i, segment := range list {
		switch segment.(type) {
		case query.Variable:
			v := segment.(query.Variable)
			return list[:i], &v, list[:i+1]
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
		out[i] = in[i].(tup.TupleElement)
	}
	return out
}
