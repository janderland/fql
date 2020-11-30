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

type QueryCtx struct {
	tr     fdb.Transaction
	ctx    context.Context
	cancel context.CancelFunc
	wgs    []*sync.WaitGroup
	errCh  chan error
}

func NewQueryCtx(tr fdb.Transaction) QueryCtx {
	ctx, cancel := context.WithCancel(context.Background())
	return QueryCtx{
		tr:     tr,
		ctx:    ctx,
		cancel: cancel,
		errCh:  make(chan error),
	}
}

func (c *QueryCtx) OpenDirectories(directory query.Directory) chan dir.DirectorySubspace {
	dirCh := make(chan dir.DirectorySubspace)

	var wg sync.WaitGroup
	c.wgs = append(c.wgs, &wg)

	wg.Add(1)
	go func() {
		defer close(dirCh)
		defer wg.Done()
		c.openDirectories(directory, dirCh)
	}()

	return dirCh
}

func (c *QueryCtx) openDirectories(directory query.Directory, dirCh chan dir.DirectorySubspace) {
	prefix, variable, suffix := splitAtVariable(directory)

	var prefixStr []string
	for _, segment := range prefix {
		prefixStr = append(prefixStr, segment.(string))
	}

	if variable != nil {
		subDirs, err := dir.List(c.tr, prefixStr)
		if err != nil {
			c.errCh <- errors.Wrap(err, "failed to list directories")
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
			c.errCh <- errors.Wrap(err, "failed to open directory")
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

func (c *QueryCtx) ReadRange(tuple query.Tuple, dirCh chan dir.DirectorySubspace) chan DirKeyValue {
	kvCh := make(chan DirKeyValue)

	var wg sync.WaitGroup
	c.wgs = append(c.wgs, &wg)

	fdbTuple := make(tup.Tuple, len(tuple))
	for i := range tuple {
		fdbTuple[i] = tuple[i].(tup.TupleElement)
	}

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

func (c *QueryCtx) readRange(tuple tup.Tuple, dirCh chan dir.DirectorySubspace, kvCh chan DirKeyValue) {
	for {
		var directory dir.DirectorySubspace
		var open bool
		select {
		case <-c.ctx.Done():
			return
		case directory, open = <-dirCh:
			if !open {
				return
			}
		}

		rng, err := fdb.PrefixRange(directory.Pack(tuple))
		if err != nil {
			select {
			case <-c.ctx.Done():
			case c.errCh <- errors.Wrap(err, "failed to create prefix range"):
			}
			return
		}

		iter := c.tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv, err := iter.Get()
			if err != nil {
				select {
				case <-c.ctx.Done():
				case c.errCh <- errors.Wrap(err, "failed to get key-value"):
				}
				return
			}
			select {
			case <-c.ctx.Done():
			case kvCh <- DirKeyValue{dir: directory, kv: kv}:
			}
		}
	}
}

func splitAtVariable(list []interface{}) ([]interface{}, *query.Variable, []interface{}) {
	for i, segment := range list {
		switch segment.(type) {
		case query.Variable:
			v := segment.(query.Variable)
			return list[:i], &v, list[:i+1]
		}
	}
	return list, nil, nil
}
