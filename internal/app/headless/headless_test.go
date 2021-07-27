package headless

import (
	"context"
	"os"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/janderland/fdbq/internal/app/flag"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHeadless_Query(t *testing.T) {
	tests := []struct {
		name  string
		flags flag.Flags
		query string
		err   bool
	}{
		{
			name:  "set",
			flags: flag.Flags{Write: true},
			query: "/my/dir(\"hi\",\"there\")=33.9",
			err:   false,
		},
		{
			name:  "set error",
			flags: flag.Flags{Write: false},
			query: "/my/dir(\"hi\",\"there\")=33.9",
			err:   true,
		},
		{
			name:  "clear",
			flags: flag.Flags{Write: true},
			query: "/my/dir(\"hi\",\"there\")=clear",
			err:   false,
		},
		{
			name:  "clear error",
			flags: flag.Flags{Write: false},
			query: "/my/dir(\"hi\",\"there\")=clear",
			err:   true,
		},
		{
			name:  "get nothing",
			flags: flag.Flags{},
			query: "/nothing/is/here{\"wont\",\"match\"}=<>",
			err:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, test.flags, func(h Headless) {
				err := h.Query(test.query)
				if test.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}

func testEnv(t *testing.T, flags flag.Flags, f func(Headless)) {
	writer := zerolog.ConsoleWriter{Out: os.Stdout}
	writer.FormatLevel = func(_ interface{}) string { return "" }
	log := zerolog.New(writer).With().Timestamp().Logger()

	dv, closeDV := devnull(t)
	defer closeDV()

	f(New(log.WithContext(context.Background()), flags, dv, &nullTransactor{}))
}

type nullTransactor struct{}

func (t *nullTransactor) Transact(_ func(fdb.Transaction) (interface{}, error)) (interface{}, error) {
	return nil, nil
}

func (t *nullTransactor) ReadTransact(_ func(fdb.ReadTransaction) (interface{}, error)) (interface{}, error) {
	return nil, nil
}

func devnull(t *testing.T) (*os.File, func()) {
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to open devnull"))
	}
	return devnull, func() {
		if err := devnull.Close(); err != nil {
			t.Fatal(errors.Wrap(err, "failed to close devnull"))
		}
	}
}
