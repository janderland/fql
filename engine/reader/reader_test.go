package reader

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

func TestReader_openDirectories(t *testing.T) {
	tests := []struct {
		name     string           // name of test
		query    keyval.Directory // query to execute
		initial  [][]string       // initial directory state
		expected [][]string       // expected results
		error    bool             // expect error?
	}{
		{
			name:  "no exist one",
			query: keyval.Directory{"hello"},
			error: true,
		},
		{
			name:     "exist one",
			query:    keyval.Directory{"hello"},
			initial:  [][]string{{"hello"}},
			expected: [][]string{{"hello"}},
		},
		{
			name:  "no exist many",
			query: keyval.Directory{"people", keyval.Variable{}},
			error: true,
		},
		{
			name:  "exist many",
			query: keyval.Directory{"people", keyval.Variable{}, "job", keyval.Variable{}},
			initial: [][]string{
				{"people", "billy", "job", "dancer"},
				{"people", "billy", "job", "tailor"},
				{"people", "jon", "job", "programmer"},
				{"people", "sally", "job", "designer"},
			},
			expected: [][]string{
				{"people", "billy", "job", "dancer"},
				{"people", "billy", "job", "tailor"},
				{"people", "jon", "job", "programmer"},
				{"people", "sally", "job", "designer"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, r Reader) {
				// Set up initial state of directories.
				for _, dir := range test.initial {
					_, err := directory.Create(tr, append([]string{root}, dir...), nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}

				// Execute the query.
				dirCh := r.openDirectories(keyval.KeyValue{
					Key: keyval.Key{Directory: append(keyval.Directory{root}, test.query...)},
				})
				waitForDirs := collectDirs(dirCh)

				// Wait for the query to complete and check for errors.
				if test.error {
					assert.Error(t, waitForErr(r))
				} else {
					assert.NoError(t, waitForErr(r))
				}

				// Collect the query output and assert it's as expected.
				directories := waitForDirs()
				if assert.Equal(t, len(test.expected), len(directories)) {
					for i := range test.expected {
						assert.Equal(t, append([]string{root}, test.expected[i]...), directories[i].GetPath())
					}
				}
			})
		})
	}
}

func testEnv(t *testing.T, f func(fdb.Transaction, Reader)) {
	exists, err := directory.Exists(db, []string{root})
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to check if root directory exists"))
	}
	if exists {
		t.Fatal(errors.New("root directory already exists"))
	}

	defer func() {
		_, err := directory.Root().Remove(db, []string{root})
		if err != nil {
			t.Error(errors.Wrap(err, "failed to clean root directory"))
		}
	}()

	_, err = db.Transact(func(tr fdb.Transaction) (interface{}, error) {
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

func waitForErr(r Reader) error {
	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	for err := range r.errCh {
		return err
	}
	return nil
}
