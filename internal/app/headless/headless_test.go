package headless

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/janderland/fql/engine"
	"github.com/janderland/fql/engine/facade"
)

func TestHeadless_Query(t *testing.T) {
	tests := []struct {
		name    string
		write   bool
		queries []string
		err     bool
	}{
		{
			name:    "set",
			write:   true,
			queries: []string{"/my/dir(\"hi\",\"there\")=33.9"},
			err:     false,
		},
		{
			name:    "set error",
			write:   false,
			queries: []string{"/my/dir(\"hi\",\"there\")=33.9"},
			err:     true,
		},
		{
			name:    "clear",
			write:   true,
			queries: []string{"/my/dir(\"hi\",\"there\")=clear"},
			err:     false,
		},
		{
			name:    "clear error",
			write:   false,
			queries: []string{"/my/dir(\"hi\",\"there\")=clear"},
			err:     true,
		},
		{
			name:    "get nothing",
			write:   false,
			queries: []string{"/nothing/is/here(\"wont\",\"match\")=<>"},
			err:     false,
		},
		{
			name:    "watch single query",
			write:   false,
			queries: []string{"/my/dir(\"test\")=<>"},
			err:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(app App) {
				app.Write = test.write

				err := app.Run(context.Background(), test.queries)
				if test.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

func TestHeadless_Watch(t *testing.T) {
	tests := []struct {
		name    string
		queries []string
		err     bool
	}{
		{
			name:    "watch single query",
			queries: []string{"/my/dir(\"test\")=<>"},
			err:     false,
		},
		{
			name:    "watch multiple queries error",
			queries: []string{"/my/dir(\"test1\")=<>", "/my/dir(\"test2\")=<>"},
			err:     true,
		},
		{
			name:    "watch range query error",
			queries: []string{"/my/dir(..)=<>"},
			err:     true,
		},
		{
			name:    "watch directory query error",
			queries: []string{"/my/dir"},
			err:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv(t, func(app App) {
				app.Watch = true

				// Use a context with timeout to prevent the watch from running forever
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				err := app.Run(ctx, test.queries)
				if test.err {
					require.Error(t, err)
				} else {
					// For successful watch tests, we expect a context timeout error
					// since the watch will keep running until cancelled
					if err != nil {
						require.Equal(t, context.DeadlineExceeded, err)
					}
				}
			})
		})
	}
}

func testEnv(t *testing.T, f func(App)) {
	writer := zerolog.ConsoleWriter{Out: os.Stdout}
	writer.FormatLevel = func(_ interface{}) string { return "" }
	writer.FormatTimestamp = func(_ interface{}) string { return "" }
	log := zerolog.New(writer)

	dv, closeDV := devnull(t)
	defer closeDV()

	f(App{
		Engine: engine.New(facade.NewNilTransactor(), engine.Logger(log)),
		Out:    dv,
	})
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
