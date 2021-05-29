package engine

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	kv "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

const root = "engine"

var (
	db fdb.Database

	flags struct {
		force bool
		level string
	}
)

func init() {
	fdb.MustAPIVersion(620)
	db = fdb.MustOpenDefault()

	flag.BoolVar(&flags.force, "force", false, "remove test directory if it exists")
	flag.StringVar(&flags.level, "level", "debug", "logging level")
}

func TestEngine_SetSingleRead(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			query := kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi", "there"}, Tuple: kv.Tuple{33.3}}, Value: int64(33)}
			query.Key.Directory = append(kv.FromStringArray(root.GetPath()), query.Key.Directory...)

			err := e.Set(query)
			assert.NoError(t, err)

			expected := query
			query.Value = kv.Variable{kv.IntType}

			result, err := e.SingleRead(query)
			assert.NoError(t, err)
			assert.Equal(t, &expected, result)
		})
	})

	t.Run("set errors", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			err := e.Set(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33, kv.Variable{}}}, Value: nil})
			assert.Error(t, err)

			err = e.Set(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33}}, Value: kv.Clear{}})
			assert.Error(t, err)
		})
	})

	t.Run("get errors", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			result, err := e.SingleRead(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33, kv.Variable{}}}, Value: nil})
			assert.Error(t, err)
			assert.Nil(t, result)

			result, err = e.SingleRead(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33}}, Value: kv.Clear{}})
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}

func TestEngine_Clear(t *testing.T) {
	t.Run("set clear get", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			set := kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"this", "place"}, Tuple: kv.Tuple{32.33}}, Value: []byte{}}
			err := e.Set(set)
			assert.NoError(t, err)

			get := set
			get.Value = kv.Variable{}
			result, err := e.SingleRead(get)
			assert.NoError(t, err)
			assert.Equal(t, &set, result)

			clear := set
			clear.Value = kv.Clear{}
			err = e.Clear(clear)
			assert.NoError(t, err)

			result, err = e.SingleRead(get)
			assert.NoError(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("errors", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			err := e.Clear(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33, kv.Variable{}}}, Value: kv.Clear{}})
			assert.Error(t, err)

			err = e.Clear(kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33}}, Value: nil})
			assert.Error(t, err)
		})
	})
}

func TestEngine_RangeRead(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			var expected []kv.KeyValue

			query := kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"place"}, Tuple: kv.Tuple{"hi you"}}, Value: nil}
			expected = append(expected, query)
			err := e.Set(query)
			assert.NoError(t, err)

			query.Key.Tuple = kv.Tuple{11.3}
			expected = append(expected, query)
			err = e.Set(query)
			assert.NoError(t, err)

			query.Key.Tuple = kv.Tuple{55.3}
			expected = append(expected, query)
			err = e.Set(query)
			assert.NoError(t, err)

			var results []kv.KeyValue
			query.Key.Tuple = kv.Tuple{kv.Variable{}}
			for kve := range e.RangeRead(context.Background(), query) {
				if !assert.NoError(t, kve.Err) {
					t.FailNow()
				}
				results = append(results, kve.KV)
			}
			assert.Equal(t, expected, results)
		})
	})

	t.Run("errors", func(t *testing.T) {
		testEnv(t, func(root directory.DirectorySubspace, e Engine) {
			query := kv.KeyValue{Key: kv.Key{Directory: kv.Directory{"hi"}, Tuple: kv.Tuple{32.33}}, Value: kv.Clear{}}
			out := e.RangeRead(context.Background(), query)

			msg := <-out
			assert.Error(t, msg.Err)
			_, open := <-out
			assert.False(t, open)
		})
	})
}

func testEnv(t *testing.T, f func(directory.DirectorySubspace, Engine)) {
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

	level, err := zerolog.ParseLevel(flags.level)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to parse logging level"))
	}
	zerolog.SetGlobalLevel(level)
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})

	f(dir, New(log.WithContext(context.Background()), db))
}
