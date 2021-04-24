package reader

import (
	"context"
	"flag"
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

var (
	db fdb.Database

	openDirectoriesTests = []struct {
		name     string           // name of test
		query    keyval.Directory // query to execute
		initial  [][]string       // initial state
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

	readRangeTests = []struct {
		name     string            // name of test
		query    keyval.Tuple      // query to execute
		initial  []keyval.KeyValue // initial state
		expected []keyval.KeyValue // expected results
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

	filterKeysTests = []struct {
		name     string            // name of test
		query    keyval.Tuple      // query to execute
		initial  []keyval.KeyValue // initial state
		expected []keyval.KeyValue // expected results
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
						Tuple:     keyval.Tuple{123, "hello", -50.6},
					},
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
						Tuple:     keyval.Tuple{123, "song", "sing"},
					},
				},
				{
					Key: keyval.Key{
						Directory: keyval.Directory{"that", "there"},
						Tuple:     keyval.Tuple{123, 13.45, "sing"},
					},
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
			expected: nil,
		},
	}

	unpackValuesTests = []struct {
		name     string            // name of test
		query    keyval.Value      // query to execute
		initial  []keyval.KeyValue // initial state
		expected []keyval.KeyValue // expected results
	}{
		{
			name:  "no variable",
			query: 123,
			initial: []keyval.KeyValue{
				{Value: packWithPanic(123)},
				{Value: packWithPanic("hello world")},
				{Value: []byte{}},
			},
			expected: []keyval.KeyValue{
				{Value: 123},
			},
		},
		{
			name:  "variable",
			query: keyval.Variable{Type: []keyval.ValueType{keyval.IntType, keyval.BigIntType, keyval.TupleType}},
			initial: []keyval.KeyValue{
				{Value: packWithPanic("hello world")},
				{Value: packWithPanic(55)},
				{Value: packWithPanic(23.9)},
				{Value: packWithPanic(keyval.Tuple{"there we go", nil})},
			},
			expected: []keyval.KeyValue{
				{Value: int64(55)},
				{Value: unpackWithPanic(keyval.IntType, packWithPanic(23.9))},
				{Value: keyval.Tuple{"there we go", nil}},
			},
		},
	}
)

func init() {
	fdb.MustAPIVersion(620)
	db = fdb.MustOpenDefault()
}

func TestReader_openDirectories(t *testing.T) {
	for _, test := range openDirectoriesTests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, r Reader) {
				// Set up initial state of directories.
				for _, path := range test.initial {
					_, err := rootDir.Create(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}

				// Execute the query.
				out := r.openDirectories(keyval.KeyValue{
					Key: keyval.Key{Directory: append(keyval.FromStringArray(rootDir.GetPath()), test.query...)},
				})
				waitForDirs := collectDirs(t, out)

				// Wait for the query to complete
				// and check for errors.
				if test.error {
					assert.Error(t, waitForErr(r))
				} else {
					assert.NoError(t, waitForErr(r))
				}

				// Collect the query output and assert it's as expected.
				directories := waitForDirs()
				if assert.Equalf(t, len(test.expected), len(directories), "unexpected number of directories") {
					for i, expected := range test.expected {
						expected = append(rootDir.GetPath(), expected...)
						if !assert.Equalf(t, expected, directories[i].GetPath(), "unexpected directory (index %d)", i) {
							t.FailNow()
						}
					}
				}
			})
		})
	}
}

func TestReader_readRange(t *testing.T) {
	for _, test := range readRangeTests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, r Reader) {
				paths := make(map[string]struct{})
				var dirs []directory.DirectorySubspace

				// Setup initial key-values.
				for _, kv := range test.initial {
					path, err := keyval.ToStringArray(kv.Key.Directory)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					dir, err := rootDir.CreateOrOpen(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					tr.Set(dir.Pack(keyval.ToFDBTuple(kv.Key.Tuple)), nil)

					pathStr := strings.Join(path, "/")
					if _, exists := paths[pathStr]; !exists {
						t.Logf("adding to dir list: %v", dir.GetPath())
						paths[pathStr] = struct{}{}
						dirs = append(dirs, dir)
					}
				}

				// ctx ensures sendDirs() returns when the test exits.
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Execute query.
				out := r.readRange(keyval.KeyValue{Key: keyval.Key{Tuple: test.query}}, sendDirs(ctx, t, dirs))
				waitForKVs := collectKVs(t, out)

				// Wait for the query to complete
				// and check for errors.
				assert.NoError(t, waitForErr(r))

				// Ensure the read key-values are as expected.
				kvs := waitForKVs()
				rootPath := keyval.FromStringArray(rootDir.GetPath())
				assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values")
				for i, expected := range test.expected {
					expected.Key.Directory = append(rootPath, expected.Key.Directory...)
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value (index %d)", i) {
						t.FailNow()
					}
				}
			})
		})
	}
}

func TestReader_filterKeys(t *testing.T) {
	for _, test := range filterKeysTests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, r Reader) {
				// ctx ensures sendKVs() returns when the test exits.
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Execute query.
				out := r.filterKeys(keyval.KeyValue{Key: keyval.Key{Tuple: test.query}}, sendKVs(ctx, t, test.initial))
				waitForKVs := collectKVs(t, out)

				// Wait for the query to complete
				// and check for errors.
				assert.NoError(t, waitForErr(r))

				// Ensure the read key-values are as expected.
				kvs := waitForKVs()
				assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values")
				for i, expected := range test.expected {
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value (index %d)", i) {
						t.FailNow()
					}
				}
			})
		})
	}
}

func TestReader_unpackValues(t *testing.T) {
	for _, test := range unpackValuesTests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, r Reader) {
				// ctx ensures sendKVs() returns when the test exits.
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Execute query.
				out := r.unpackValues(keyval.KeyValue{Value: test.query}, sendKVs(ctx, t, test.initial))
				waitForKVs := collectKVs(t, out)

				// Wait for the query to complete
				// and check for errors.
				assert.NoError(t, waitForErr(r))

				// Ensure the read key-values are as expected.
				kvs := waitForKVs()
				assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values")
				for i, expected := range test.expected {
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value (index %d)", i) {
						t.FailNow()
					}
				}
			})
		})
	}
}

var allowedToDelete = flag.Bool("force", false, "remove test directory if it exists")

func testEnv(t *testing.T, f func(fdb.Transaction, directory.DirectorySubspace, Reader)) {
	exists, err := directory.Exists(db, []string{root})
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to check if root directory exists"))
	}
	if exists {
		if !*allowedToDelete {
			t.Fatal(errors.New("test directory already exists, use '-force' flag to remove"))
		}
		if _, err := directory.Root().Remove(db, []string{root}); err != nil {
			t.Fatal(errors.Wrap(err, "failed to remove directory"))
		}
	}

	dir, err := directory.Create(db, []string{root}, nil)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to create test directory"))
	}
	defer func() {
		_, err := directory.Root().Remove(db, []string{root})
		if err != nil {
			t.Error(errors.Wrap(err, "failed to clean root directory"))
		}
	}()

	_, err = db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		f(tr, dir, New(context.Background(), tr))
		return nil, nil
	})
	if err != nil {
		t.Fatal(errors.Wrap(err, "transaction failed"))
	}
}

func collectDirs(t *testing.T, in chan directory.DirectorySubspace) func() []directory.DirectorySubspace {
	var out []directory.DirectorySubspace
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for dir := range in {
			t.Logf("received directory: %s", dir.GetPath())
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

func sendKVs(ctx context.Context, t *testing.T, in []keyval.KeyValue) chan keyval.KeyValue {
	out := make(chan keyval.KeyValue)

	go func() {
		defer close(out)
		for _, kv := range in {
			select {
			case <-ctx.Done():
				return
			case out <- kv:
				t.Logf("sent kv: %+v", kv)
			}
		}
	}()

	return out
}

func packWithPanic(val keyval.Value) []byte {
	packed, err := keyval.PackValue(val)
	if err != nil {
		panic(err)
	}
	return packed
}

func unpackWithPanic(typ keyval.ValueType, bytes []byte) keyval.Value {
	unpacked, err := keyval.UnpackValue(typ, bytes)
	if err != nil {
		panic(err)
	}
	return unpacked
}

func waitForErr(r Reader) error {
	go func() {
		r.wg.Wait()
		close(r.errCh)
	}()

	return <-r.errCh
}
