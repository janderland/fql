package stream

import (
	"context"
	"encoding/binary"
	"flag"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/keyval/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const root = "root"

var (
	db        fdb.Database
	byteOrder binary.ByteOrder

	flags struct {
		force bool
	}
)

func init() {
	fdb.MustAPIVersion(620)
	db = fdb.MustOpenDefault()
	byteOrder = binary.BigEndian

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
			query: q.Directory{q.String("hello")},
			error: true,
		},
		{
			name:     "exist one",
			query:    q.Directory{q.String("hello")},
			initial:  [][]string{{"hello"}},
			expected: [][]string{{"hello"}},
		},
		{
			name:  "no exist many",
			query: q.Directory{q.String("people"), q.Variable{}},
			error: true,
		},
		{
			name:  "exist many",
			query: q.Directory{q.String("people"), q.Variable{}, q.String("job"), q.Variable{}},
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

				out := s.OpenDirectories(tr, append(convert.FromStringArray(rootDir.GetPath()), test.query...))
				directories, err := collectDirs(out)
				if test.error {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				if !assert.Equalf(t, len(test.expected), len(directories), "unexpected number of directories") {
					t.FailNow()
				}
				for i, expected := range test.expected {
					expected = append(rootDir.GetPath(), expected...)
					if !assert.Equalf(t, expected, directories[i].GetPath(), "unexpected directory at index %d", i) {
						t.FailNow()
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
			query: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)}}},
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(321), q.String("goodbye"), q.Float(50.6)}}},
				{Key: q.Key{Directory: q.Directory{q.String("second")}, Tuple: q.Tuple{q.Int(-69), q.BigInt(*big.NewInt(-55)), q.Tuple{q.String("world")}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)}}, Value: q.Bytes{}},
			},
		},
		{
			name:  "variable",
			query: q.Tuple{q.Int(123), q.Variable{}, q.String("sing")},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("iam")}, Tuple: q.Tuple{q.UUID{
					0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}, Value: q.Bytes{}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}, Value: q.Bytes{}},
			},
		},
		{
			name:  "read everything",
			query: q.Tuple{},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("iam")}, Tuple: q.Tuple{
					q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}, Value: q.Bytes{}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}, Value: q.Bytes{}},
				{Key: q.Key{Directory: q.Directory{q.String("iam")}, Tuple: q.Tuple{
					q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}, Value: q.Bytes{}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				var dirs []directory.DirectorySubspace
				paths := make(map[string]struct{})

				for _, kv := range test.initial {
					path, err := convert.ToStringArray(kv.Key.Directory)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					tup, err := convert.ToFDBTuple(kv.Key.Tuple)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					dir, err := rootDir.CreateOrOpen(tr, path, nil)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					tr.Set(dir.Pack(tup), nil)

					pathStr := strings.Join(path, "/")
					if _, exists := paths[pathStr]; !exists {
						t.Logf("adding to dir list: %v", path)
						paths[pathStr] = struct{}{}
						dirs = append(dirs, dir)
					}
				}

				out := s.ReadRange(tr, test.query, RangeOpts{}, sendDirs(t, s, dirs))
				kvs, err := collectKVs(out)
				assert.NoError(t, err)

				rootPath := convert.FromStringArray(rootDir.GetPath())
				if !assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values") {
					t.FailNow()
				}
				for i, expected := range test.expected {
					expected.Key.Directory = append(rootPath, expected.Key.Directory...)
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value at index %d", i) {
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
			query: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)}}},
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(321), q.String("goodbye"), q.Float(50.6)}}},
				{Key: q.Key{Directory: q.Directory{q.String("second")}, Tuple: q.Tuple{q.Int(-69), q.BigInt(*big.NewInt(-55)), q.Tuple{q.String("world")}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("first")}, Tuple: q.Tuple{q.Int(123), q.String("hello"), q.Float(-50.6)}}},
			},
		},
		{
			name:  "variable",
			query: q.Tuple{q.Int(123), q.Variable{}, q.String("sing")},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("iam")}, Tuple: q.Tuple{q.UUID{
					0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}},
			},
		},
		{
			name:  "read everything",
			query: q.Tuple{},
			initial: []q.KeyValue{
				{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("thing")}, Tuple: q.Tuple{q.Int(123), q.String("song"), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("that"), q.String("there")}, Tuple: q.Tuple{q.Int(123), q.Float(13.45), q.String("sing")}}},
				{Key: q.Key{Directory: q.Directory{q.String("iam")}, Tuple: q.Tuple{
					q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				out := s.FilterKeys(test.query, sendKVs(t, s, test.initial))
				kvs, err := collectKVs(out)
				assert.NoError(t, err)

				if !assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values") {
					t.FailNow()
				}
				for i, expected := range test.expected {
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value at index %d", i) {
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
			query: q.Int(123),
			initial: []q.KeyValue{
				{Value: packWithPanic(q.Int(123))},
				{Value: packWithPanic(q.String("hello world"))},
				{Value: q.Bytes{}},
			},
			expected: []q.KeyValue{
				{Value: q.Int(123)},
			},
		},
		{
			name:  "variable",
			query: q.Variable{q.IntType, q.BigIntType, q.TupleType},
			initial: []q.KeyValue{
				{Value: packWithPanic(q.String("hello world"))},
				{Value: packWithPanic(q.Int(55))},
				{Value: packWithPanic(q.Float(23.9))},
				{Value: packWithPanic(q.Tuple{q.String("there we go"), q.Nil{}})},
			},
			expected: []q.KeyValue{
				{Value: q.Int(55)},
				{Value: unpackWithPanic(q.IntType, packWithPanic(q.Float(23.9)))},
				{Value: q.Tuple{q.String("there we go"), q.Nil{}}},
			},
		},
		{
			name:  "empty variable",
			query: q.Variable{},
			initial: []q.KeyValue{
				{Value: packWithPanic(q.Int(55))},
				{Value: packWithPanic(q.Float(23.9))},
				{Value: packWithPanic(q.Tuple{q.String("there we go"), q.Nil{}})},
			},
			expected: []q.KeyValue{
				{Value: packWithPanic(q.Int(55))},
				{Value: packWithPanic(q.Float(23.9))},
				{Value: packWithPanic(q.Tuple{q.String("there we go"), q.Nil{}})},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(tr fdb.Transaction, rootDir directory.DirectorySubspace, s Stream) {
				out := s.UnpackValues(test.query, byteOrder, sendKVs(t, s, test.initial))
				kvs, err := collectKVs(out)
				assert.NoError(t, err)

				if !assert.Equal(t, len(test.expected), len(kvs), "unexpected number of key-values") {
					t.FailNow()
				}
				for i, expected := range test.expected {
					if !assert.Equalf(t, expected, kvs[i], "unexpected key-value at index %d", i) {
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

func packWithPanic(val q.Value) q.Bytes {
	packed, err := values.Pack(val, byteOrder)
	if err != nil {
		panic(err)
	}
	return packed
}

func unpackWithPanic(typ q.ValueType, bytes q.Bytes) q.Value {
	unpacked, err := values.Unpack(bytes, typ, byteOrder)
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
