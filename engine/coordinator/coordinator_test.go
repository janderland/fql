package coordinator

import (
	"sync"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const root = "root"

var db fdb.Database

func init() {
	fdb.MustAPIVersion(610)
	db = fdb.MustOpenDefault()
}

func TestCoordinator_OpenDirectories(t *testing.T) {
	t.Run("no exist one", func(t *testing.T) {
		test(t, func(_ fdb.Transaction, c Coordinator) {
			dirCh := c.OpenDirectories(keyval.Directory{root, "hello"}, false)
			_ = collectDirs(dirCh) // unblock sender
			assert.Error(t, c.Wait())
		})
	})

	t.Run("exists one", func(t *testing.T) {
		test(t, func(tr fdb.Transaction, c Coordinator) {
			_, err := directory.Create(tr, []string{root, "hello"}, nil)
			assert.NoError(t, err)

			dirCh := c.OpenDirectories(keyval.Directory{root, "hello"}, false)
			waitForDirs := collectDirs(dirCh)

			assert.NoError(t, c.Wait())
			directories := waitForDirs()
			assert.Equal(t, 1, len(directories))
			assert.Equal(t, []string{root, "hello"}, directories[0].GetPath())
		})
	})

	t.Run("create one", func(t *testing.T) {
		test(t, func(tr fdb.Transaction, c Coordinator) {
			dirCh := c.OpenDirectories(keyval.Directory{root, "what", "who"}, true)
			waitForDirs := collectDirs(dirCh)

			assert.NoError(t, c.Wait())
			directories := waitForDirs()
			assert.Equal(t, 1, len(directories))
			assert.Equal(t, []string{root, "what", "who"}, directories[0].GetPath())
		})
	})
}

func test(t *testing.T, f func(fdb.Transaction, Coordinator)) {
	defer func() {
		_, err := directory.Root().Remove(db, []string{root})
		if err != nil {
			t.Error(errors.Wrap(err, "failed to clean root directory"))
		}
	}()

	_, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		f(tr, New(tr))
		return nil, nil
	})
	if err != nil {
		t.Error(errors.Wrap(err, "transaction failed"))
	}
}

func collectDirs(dirCh chan directory.DirectorySubspace) func() []directory.DirectorySubspace {
	var directories []directory.DirectorySubspace
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for dir := range dirCh {
			directories = append(directories, dir)
		}
	}()

	return func() []directory.DirectorySubspace {
		wg.Wait()
		return directories
	}
}
