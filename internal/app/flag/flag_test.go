package flag

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		dv, closeDV := devnull(t)
		defer closeDV()

		expectedFlags := Flags{
			Write:   true,
			Cluster: "my.cluster",
		}

		expectedQueries := []string{"query", "request"}

		args := []string{"", "-cluster", "my.cluster", "-write", expectedQueries[0], expectedQueries[1]}
		flags, queries, err := Parse(args, dv)

		require.NoError(t, err)
		require.Equal(t, &expectedFlags, flags)
		require.Equal(t, expectedQueries, queries)
	})

	t.Run("help", func(t *testing.T) {
		dv, closeDV := devnull(t)
		defer closeDV()

		flags, queries, err := Parse([]string{"", "-h"}, dv)
		require.NoError(t, err)
		require.Nil(t, flags)
		require.Nil(t, queries)
	})

	t.Run("failure", func(t *testing.T) {
		dv, closeDV := devnull(t)
		defer closeDV()

		flags, queries, err := Parse([]string{"", "-bad"}, dv)
		require.Error(t, err)
		require.Nil(t, flags)
		require.Nil(t, queries)
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
