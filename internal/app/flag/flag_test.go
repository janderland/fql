package flag

import (
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
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

		assert.NoError(t, err)
		assert.Equal(t, &expectedFlags, flags)
		assert.Equal(t, expectedQueries, queries)
	})

	t.Run("help", func(t *testing.T) {
		dv, closeDV := devnull(t)
		defer closeDV()

		flags, queries, err := Parse([]string{"", "-h"}, dv)
		assert.NoError(t, err)
		assert.Nil(t, flags)
		assert.Nil(t, queries)
	})

	t.Run("failure", func(t *testing.T) {
		dv, closeDV := devnull(t)
		defer closeDV()

		flags, queries, err := Parse([]string{"", "-bad"}, dv)
		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Nil(t, queries)
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
