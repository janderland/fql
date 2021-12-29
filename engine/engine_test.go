package engine

import (
	"context"
	"encoding/binary"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/janderland/fdbq/engine/facade"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const root = "engine"

var (
	byteOrder binary.ByteOrder

	flags struct {
		force bool
	}
)

func init() {
	fdb.MustAPIVersion(620)
	byteOrder = binary.BigEndian

	flag.BoolVar(&flags.force, "force", false, "remove test directory if it exists")
}

func TestEngine_SetSingleRead(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi"), q.String("there")}, Tuple: q.Tuple{q.Float(33.3)}}, Value: q.Int(33)})
			err := e.Set(query, byteOrder)
			require.NoError(t, err)

			expected := query
			query.Value = q.Variable{q.IntType}
			result, err := e.SingleRead(query, SingleOpts{ByteOrder: byteOrder})
			require.NoError(t, err)
			require.Equal(t, &expected, result)
		})
	})

	t.Run("set and get empty value", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi"), q.String("there")}, Tuple: q.Tuple{q.Float(33.3)}}, Value: q.Bytes{}})
			err := e.Set(query, byteOrder)
			require.NoError(t, err)

			expected := query
			query.Value = q.Variable{}
			result, err := e.SingleRead(query, SingleOpts{ByteOrder: byteOrder})
			require.NoError(t, err)
			require.Equal(t, &expected, result)
		})
	})

	t.Run("get nothing", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("nothing"), q.String("here")}}, Value: q.Variable{}})
			result, err := e.SingleRead(query, SingleOpts{ByteOrder: byteOrder})
			require.NoError(t, err)
			require.Nil(t, result)
		})
	})

	t.Run("set errors", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33), q.Variable{}}}, Value: q.Nil{}})
			err := e.Set(query, byteOrder)
			require.Error(t, err)

			query = prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33)}}, Value: q.Clear{}})
			err = e.Set(query, byteOrder)
			require.Error(t, err)
		})
	})

	t.Run("get errors", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33), q.Variable{}}}, Value: q.Nil{}})
			result, err := e.SingleRead(query, SingleOpts{ByteOrder: byteOrder})
			require.Error(t, err)
			require.Nil(t, result)

			query = prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33)}}, Value: q.Clear{}})
			result, err = e.SingleRead(query, SingleOpts{ByteOrder: byteOrder})
			require.Error(t, err)
			require.Nil(t, result)
		})
	})
}

func TestEngine_Clear(t *testing.T) {
	t.Run("set clear get", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			set := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("this"), q.String("place")}, Tuple: q.Tuple{q.Float(32.33)}}, Value: q.Bytes{}})
			err := e.Set(set, byteOrder)
			require.NoError(t, err)

			get := set
			get.Value = q.Variable{}
			result, err := e.SingleRead(get, SingleOpts{ByteOrder: byteOrder})
			require.NoError(t, err)
			require.Equal(t, &set, result)

			clear := set
			clear.Value = q.Clear{}
			err = e.Clear(clear)
			require.NoError(t, err)

			result, err = e.SingleRead(get, SingleOpts{ByteOrder: byteOrder})
			require.NoError(t, err)
			require.Nil(t, result)
		})
	})

	t.Run("errors", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33), q.Variable{}}}, Value: q.Clear{}})
			err := e.Clear(query)
			require.Error(t, err)

			query = prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33)}}, Value: q.Nil{}})
			err = e.Clear(query)
			require.Error(t, err)
		})
	})
}

func TestEngine_RangeRead(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			var expected []q.KeyValue

			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("place")}, Tuple: q.Tuple{q.String("hi you")}}, Value: q.Bytes{}})
			expected = append(expected, query)
			err := e.Set(query, byteOrder)
			require.NoError(t, err)

			query.Key.Tuple = q.Tuple{q.Float(11.3)}
			expected = append(expected, query)
			err = e.Set(query, byteOrder)
			require.NoError(t, err)

			query.Key.Tuple = q.Tuple{q.Float(55.3)}
			expected = append(expected, query)
			err = e.Set(query, byteOrder)
			require.NoError(t, err)

			var results []q.KeyValue
			query.Key.Tuple = q.Tuple{q.Variable{}}
			for kve := range e.RangeRead(context.Background(), query, RangeOpts{}) {
				require.NoError(t, kve.Err)
				results = append(results, kve.KV)
			}
			require.Equal(t, expected, results)
		})
	})

	t.Run("errors", func(t *testing.T) {
		testEnv(t, func(_ fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			query := prefixDir(root, q.KeyValue{Key: q.Key{Directory: q.Directory{q.String("hi")}, Tuple: q.Tuple{q.Float(32.33)}}, Value: q.Clear{}})
			out := e.RangeRead(context.Background(), query, RangeOpts{ByteOrder: byteOrder})

			msg := <-out
			require.Error(t, msg.Err)
			_, open := <-out
			require.False(t, open)
		})
	})
}

func TestEngine_Directories(t *testing.T) {
	t.Run("created and open", func(t *testing.T) {
		testEnv(t, func(tr fdb.Transactor, root directory.DirectorySubspace, e Engine) {
			paths := [][]string{
				{"my", "path"},
				{"my", "somewhere"},
			}

			var expected []directory.DirectorySubspace
			for _, p := range paths {
				dir, err := root.Create(tr, p, nil)
				require.NoError(t, err)
				expected = append(expected, dir)
			}

			var result []directory.DirectorySubspace
			query := append(convert.FromStringArray(root.GetPath()), q.Directory{q.String("my"), q.Variable{}}...)
			for msg := range e.Directories(context.Background(), query) {
				require.NoError(t, msg.Err)
				result = append(result, msg.Dir)
			}
			require.Equal(t, expected, result)
		})
	})
}

func testEnv(t *testing.T, f func(fdb.Transactor, directory.DirectorySubspace, Engine)) {
	db := fdb.MustOpenDefault()
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
	writer.FormatTimestamp = func(_ interface{}) string { return "" }
	log := zerolog.New(writer)

	f(db, dir, Engine{Tr: facade.NewTransactor(db), Log: log})
}

func prefixDir(root directory.DirectorySubspace, query q.KeyValue) q.KeyValue {
	query.Key.Directory = append(convert.FromStringArray(root.GetPath()), query.Key.Directory...)
	return query
}
