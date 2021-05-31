package stream

import (
	"context"
	"flag"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const root = "root"

var (
	db fdb.Database

	flags struct {
		force bool
	}
)

func init() {
	fdb.MustAPIVersion(620)
	db = fdb.MustOpenDefault()

	flag.BoolVar(&flags.force, "force", false, "remove test directory if it exists")
}

func TestStream_OpenDirectories(t *testing.T) {
	var tests = []struct {
		name     string
		query    q.Directory
		initial  [][]string
		expected [][]string
		error    bool
	}{
		{
			name:  "no exist one",
			query: q.Directory{"hello"},
			error: true,
		},
		{
			name:     "exist one",
			query:    q.Directory{"hello"},
			initial:  [][]string{{"hello"}},
			expected: [][]string{{"hello"}},
		},
		{
			name:  "no exist many",
			query: q.Directory{"people", q.Variable{}},
			error: true,
		},
		{
			name:  "exist many",
			query: q.Directory{"people", q.Variable{}, "job", q.Variable{}},
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
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				for _, path := range test.initial {
					_, err := rootDir.Create(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}

				out := s.OpenDirectories(tr, q.KeyValue{
					Key: q.Key{Directory: append(q.FromStringArray(rootDir.GetPath()), test.query...)},
				})

				directories, err := collectDirs(out)
				if test.error {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

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

func TestStream_ReadRange(t *testing.T) {
	var tests = []struct {
		name     string
		query    q.Tuple
		initial  []q.KeyValue
		expected []q.KeyValue
	}{
		{
			name:  "no variable",
			query: q.Tuple{123, "hello", -50.6},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{123, "hello", -50.6}}},
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{321, "goodbye", 50.6}}},
				{Key: q.Key{Directory: q.Directory{"second"}, Tuple: q.Tuple{-69, big.NewInt(-55), tuple.Tuple{"world"}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{int64(123), "hello", -50.6}}, Value: []byte{}},
			},
		},
		{
			name:  "variable",
			query: q.Tuple{123, q.Variable{}, "sing"},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{123, "song", "sing"}}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{123, 13.45, "sing"}}},
				{Key: q.Key{Directory: q.Directory{"iam"}, Tuple: q.Tuple{tuple.UUID{
					0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{int64(123), "song", "sing"}}, Value: []byte{}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{int64(123), 13.45, "sing"}}, Value: []byte{}},
			},
		},
		{
			name:  "read everything",
			query: q.Tuple{},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{123, "song", "sing"}}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{123, 13.45, "sing"}}},
				{Key: q.Key{Directory: q.Directory{"iam"}, Tuple: q.Tuple{
					tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{int64(123), "song", "sing"}}, Value: []byte{}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{int64(123), 13.45, "sing"}}, Value: []byte{}},
				{Key: q.Key{Directory: q.Directory{"iam"}, Tuple: q.Tuple{
					tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}, Value: []byte{}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				paths := make(map[string]struct{})
				var dirs []directory.DirectorySubspace

				for _, kv := range test.initial {
					path, err := q.ToStringArray(kv.Key.Directory)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					dir, err := rootDir.CreateOrOpen(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					tr.Set(dir.Pack(q.ToFDBTuple(kv.Key.Tuple)), nil)

					pathStr := strings.Join(path, "/")
					if _, exists := paths[pathStr]; !exists {
						t.Logf("adding to dir list: %v", dir.GetPath())
						paths[pathStr] = struct{}{}
						dirs = append(dirs, dir)
					}
				}

				out := s.ReadRange(tr, q.KeyValue{Key: q.Key{Tuple: test.query}}, sendDirs(t, s, dirs))

				kvs, err := collectKVs(out)
				assert.NoError(t, err)

				rootPath := q.FromStringArray(rootDir.GetPath())
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

func TestStream_FilterKeys(t *testing.T) {
	var tests = []struct {
		name     string
		query    q.Tuple
		initial  []q.KeyValue
		expected []q.KeyValue
	}{
		{
			name:  "no variable",
			query: q.Tuple{123, "hello", -50.6},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{123, "hello", -50.6}}},
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{321, "goodbye", 50.6}}},
				{Key: q.Key{Directory: q.Directory{"second"}, Tuple: q.Tuple{-69, big.NewInt(-55), tuple.Tuple{"world"}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"first"}, Tuple: q.Tuple{123, "hello", -50.6}}},
			},
		},
		{
			name:  "variable",
			query: q.Tuple{123, q.Variable{}, "sing"},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{123, "song", "sing"}}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{123, 13.45, "sing"}}},
				{Key: q.Key{Directory: q.Directory{"iam"}, Tuple: q.Tuple{tuple.UUID{
					0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{123, "song", "sing"}}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{123, 13.45, "sing"}}},
			},
		},
		{
			name:  "read everything",
			query: q.Tuple{},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{"this", "thing"}, Tuple: q.Tuple{123, "song", "sing"}}},
				{Key: q.Key{Directory: q.Directory{"that", "there"}, Tuple: q.Tuple{123, 13.45, "sing"}}},
				{Key: q.Key{Directory: q.Directory{"iam"}, Tuple: q.Tuple{
					tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				out := s.FilterKeys(q.KeyValue{Key: q.Key{Tuple: test.query}}, sendKVs(t, s, test.initial))

				kvs, err := collectKVs(out)
				assert.NoError(t, err)

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

func TestStream_UnpackValues(t *testing.T) {
	var tests = []struct {
		name     string
		query    q.Value
		initial  []q.KeyValue
		expected []q.KeyValue
	}{
		{
			name:  "no variable",
			query: 123,
			initial: []q.KeyValue{
				{Value: packWithPanic(123)},
				{Value: packWithPanic("hello world")},
				{Value: []byte{}},
			},
			expected: []q.KeyValue{
				{Value: 123},
			},
		},
		{
			name:  "variable",
			query: q.Variable{q.IntType, q.BigIntType, q.TupleType},
			initial: []q.KeyValue{
				{Value: packWithPanic("hello world")},
				{Value: packWithPanic(55)},
				{Value: packWithPanic(23.9)},
				{Value: packWithPanic(q.Tuple{"there we go", nil})},
			},
			expected: []q.KeyValue{
				{Value: int64(55)},
				{Value: unpackWithPanic(q.IntType, packWithPanic(23.9))},
				{Value: q.Tuple{"there we go", nil}},
			},
		},
		{
			name:  "empty variable",
			query: q.Variable{},
			initial: []q.KeyValue{
				{Value: packWithPanic(55)},
				{Value: packWithPanic(23.9)},
				{Value: packWithPanic(q.Tuple{"there we go", nil})},
			},
			expected: []q.KeyValue{
				{Value: packWithPanic(55)},
				{Value: packWithPanic(23.9)},
				{Value: packWithPanic(q.Tuple{"there we go", nil})},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				out := s.UnpackValues(q.KeyValue{Value: test.query}, sendKVs(t, s, test.initial))

				kvs, err := collectKVs(out)
				assert.NoError(t, err)

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

func testEnv(t *testing.T, f func(fdb.Transaction, directory.DirectorySubspace, Stream)) {
	exists, err := directory.Exists(db, []string{root})
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to check if root directory exists"))
	}
	if exists {
		if !flags.force {
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

	writer := zerolog.ConsoleWriter{Out: os.Stdout}
	writer.FormatLevel = func(_ interface{}) string { return "" }
	writer.FormatTimestamp = func(i interface{}) string { return "" }
	log := zerolog.New(writer)

	_, err = db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		s, stop := New(log.WithContext(context.Background()))
		defer stop()

		f(tr, dir, s)
		return nil, nil
	})
	if err != nil {
		t.Fatal(errors.Wrap(err, "transaction failed"))
	}
}

func packWithPanic(val q.Value) []byte {
	packed, err := q.PackValue(val)
	if err != nil {
		panic(err)
	}
	return packed
}

func unpackWithPanic(typ q.ValueType, bytes []byte) q.Value {
	unpacked, err := q.UnpackValue(typ, bytes)
	if err != nil {
		panic(err)
	}
	return unpacked
}

func collectDirs(in chan DirErr) ([]directory.DirectorySubspace, error) {
	var out []directory.DirectorySubspace
	for msg := range in {
		if msg.Err != nil {
			return nil, msg.Err
		}
		out = append(out, msg.Dir)
	}
	return out, nil
}

func collectKVs(in chan KeyValErr) ([]q.KeyValue, error) {
	var out []q.KeyValue
	for msg := range in {
		if msg.Err != nil {
			return nil, msg.Err
		}
		out = append(out, msg.KV)
	}
	return out, nil
}

func sendDirs(t *testing.T, s Stream, in []directory.DirectorySubspace) chan DirErr {
	out := make(chan DirErr)

	go func() {
		defer close(out)
		for _, dir := range in {
			if !s.SendDir(out, DirErr{Dir: dir}) {
				return
			}
			t.Logf("sent dir: %s", dir.GetPath())
		}
	}()

	return out
}

func sendKVs(t *testing.T, s Stream, in []q.KeyValue) chan KeyValErr {
	out := make(chan KeyValErr)

	go func() {
		defer close(out)
		for _, kv := range in {
			if !s.SendKV(out, KeyValErr{KV: kv}) {
				return
			}
			t.Logf("sent kv: %+v", kv)
		}
	}()

	return out
}
