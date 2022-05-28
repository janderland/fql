package headless

import (
	"context"
	"os"
	"testing"

	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/internal/app/flag"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestHeadless_Query(t *testing.T) {
	tests := []struct {
		name    string
		flags   flag.Flags
		queries []string
		err     bool
	}{
		{
			name:    "set",
			flags:   flag.Flags{Write: true},
			queries: []string{"/my/dir{\"hi\",\"there\"}=33.9"},
			err:     false,
		},
		{
			name:    "set error",
			flags:   flag.Flags{Write: false},
			queries: []string{"/my/dir{\"hi\",\"there\"}=33.9"},
			err:     true,
		},
		{
			name:    "clear",
			flags:   flag.Flags{Write: true},
			queries: []string{"/my/dir{\"hi\",\"there\"}=clear"},
			err:     false,
		},
		{
			name:    "clear error",
			flags:   flag.Flags{Write: false},
			queries: []string{"/my/dir{\"hi\",\"there\"}=clear"},
			err:     true,
		},
		{
			name:    "get nothing",
			flags:   flag.Flags{},
			queries: []string{"/nothing/is/here{\"wont\",\"match\"}=<>"},
			err:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, test.flags, func(h App, tr facade.Transactor) {
				err := h.Run(context.Background(), tr, test.queries)
				if test.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

func testEnv(t *testing.T, flags flag.Flags, f func(App, facade.Transactor)) {
	writer := zerolog.ConsoleWriter{Out: os.Stdout}
	writer.FormatLevel = func(_ interface{}) string { return "" }
	writer.FormatTimestamp = func(_ interface{}) string { return "" }
	log := zerolog.New(writer)

	dv, closeDV := devnull(t)
	defer closeDV()

	f(App{
		Flags: flags,
		Log:   log,
		Out:   dv,
	}, facade.NewNilTransactor())
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
