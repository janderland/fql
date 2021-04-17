package reader

import (
	"context"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
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
				for _, path := range test.initial {
					_, err := directory.Create(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}

				// Execute the query.
				out := r.openDirectories(keyval.KeyValue{
					Key: keyval.Key{Directory: append(keyval.Directory{root}, test.query...)},
				})
				waitForDirs := collectDirs(out)

				// Wait for the query to complete
				// and check for errors.
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

func TestReader_readRange(t *testing.T) {
	tests := []struct {
		name     string
		query    keyval.Tuple
		initial  []keyval.KeyValue
		expected []keyval.KeyValue
	}{
		{
			name:  "no variable",
			query: keyval.Tuple{123, "hello", -50.6},
			initial: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"first"},
						Tuple:     keyval.Tuple{123, "hello", -50.6},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"first"},
						Tuple:     keyval.Tuple{321, "goodbye", 50.6},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"second"},
						Tuple:     keyval.Tuple{-69, big.NewInt(-55), tuple.Tuple{"world"}},
					},
				},
			},
			expected: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"first"},
						Tuple:     keyval.Tuple{int64(123), "hello", -50.6},
					},
					Value: []byte{},
				},
			},
		},

		{
			name:  "variable",
			query: keyval.Tuple{123, keyval.Variable{}, "sing"},
			initial: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"this", "thing"},
						Tuple:     keyval.Tuple{123, "song", "sing"},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"that", "there"},
						Tuple:     keyval.Tuple{123, 13.45, "sing"},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"iam"},
						Tuple:     keyval.Tuple{tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}},
					},
				},
			},
			expected: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"this", "thing"},
						Tuple:     keyval.Tuple{int64(123), "song", "sing"},
					},
					Value: []byte{},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"that", "there"},
						Tuple:     keyval.Tuple{int64(123), 13.45, "sing"},
					},
					Value: []byte{},
				},
			},
		},

		{
			name:  "read everything",
			query: keyval.Tuple{},
			initial: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"this", "thing"},
						Tuple:     keyval.Tuple{123, "song", "sing"},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"that", "there"},
						Tuple:     keyval.Tuple{123, 13.45, "sing"},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"iam"},
						Tuple:     keyval.Tuple{tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}},
					},
				},
			},
			expected: []keyval.KeyValue{
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"this", "thing"},
						Tuple:     keyval.Tuple{int64(123), "song", "sing"},
					},
					Value: []byte{},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"that", "there"},
						Tuple:     keyval.Tuple{int64(123), 13.45, "sing"},
					},
					Value: []byte{},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"iam"},
						Tuple:     keyval.Tuple{tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}},
					},
					Value: []byte{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, r Reader) {
				paths := make(map[string]struct{})
				var dirs []directory.DirectorySubspace

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Setup initial key-values.
				for _, kv := range test.initial {
					path, err := keyval.ToStringArray(kv.Key.Directory)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					dir, err := directory.CreateOrOpen(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					tr.Set(dir.Pack(keyval.ToFDBTuple(kv.Key.Tuple)), nil)

					pathStr := strings.Join(path, "/")
					if _, exists := paths[pathStr]; !exists {
						t.Logf("adding to dir list: %s", pathStr)
						paths[pathStr] = struct{}{}
						dirs = append(dirs, dir)
					}
				}

				// Execute query.
				out := r.readRange(keyval.KeyValue{Key: keyval.Key{Tuple: test.query}}, sendDirs(ctx, t, dirs))
				waitForKVs := collectKVs(t, out)

				// Wait for the query to complete
				// and check for errors.
				assert.NoError(t, waitForErr(r))

				// Ensure the read key-values are as expected.
				assert.Equal(t, test.expected, waitForKVs())
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

func collectDirs(in chan directory.DirectorySubspace) func() []directory.DirectorySubspace {
	var out []directory.DirectorySubspace
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for dir := range in {
			out = append(out, dir)
		}
	}()

	return func() []directory.DirectorySubspace {
		wg.Wait()
		return out
	}
}

func collectKVs(t *testing.T, in chan keyval.KeyValue) func() []keyval.KeyValue {
	var out []keyval.KeyValue
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for kv := range in {
			t.Logf("received kv: %+v", kv)
			out = append(out, kv)
		}
	}()

	return func() []keyval.KeyValue {
		wg.Wait()
		return out
	}
}

func sendDirs(ctx context.Context, t *testing.T, in []directory.DirectorySubspace) chan directory.DirectorySubspace {
	out := make(chan directory.DirectorySubspace)

	go func() {
		defer close(out)
		for _, dir := range in {
			select {
			case <-ctx.Done():
				return
			case out <- dir:
				t.Logf("sent dir: %s", dir.GetPath())
			}
		}
	}()

	return out
}

func sendKVs(ctx context.Context, in []keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)

	go func() {
		defer close(out)
		for _, kv := range in {
			select {
			case <-ctx.Done():
				return
			case out <- kv:
			}
		}
	}()

	return out
}

func waitForErr(r Reader) error {
	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	return <-r.errCh
}
