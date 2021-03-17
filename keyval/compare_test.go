package keyval

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareTuples(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		candidate := Tuple{
			int64(-8742),
			uint64(12342),
			"goodbye world",
			float64(-55.93),
			true,
			nil,
			big.NewInt(math.MaxInt64),
		}
		pattern := make(Tuple, len(candidate))
		assert.Equal(t, copy(pattern, candidate), len(candidate))

		mismatch, err := CompareTuples(pattern, candidate)
		assert.NoError(t, err)
		assert.Empty(t, mismatch)
	})

	t.Run("variable", func(t *testing.T) {
		candidate1 := Tuple{
			-8742,
			uint64(12342),
			"goodbye world",
			nil,
		}
		candidate2 := Tuple{
			-8742,
			Tuple{"nowhere", big.NewInt(55)},
			"goodbye world",
			555,
		}
		pattern := Tuple{
			-8742,
			Variable{},
			"goodbye world",
			Variable{},
		}

		mismatch, err := CompareTuples(pattern, candidate1)
		assert.NoError(t, err)
		assert.Empty(t, mismatch)

		mismatch, err = CompareTuples(pattern, candidate2)
		assert.NoError(t, err)
		assert.Empty(t, mismatch)
	})

	t.Run("too long", func(t *testing.T) {
		candidate := Tuple{
			int64(-8742),
			uint64(12342),
			"goodbye world",
			nil,
			Tuple{-55.1428},
		}
		pattern := Tuple{
			int64(-8742),
			Variable{},
			"goodbye world",
			Variable{},
		}

		mismatch, err := CompareTuples(pattern, candidate)
		assert.NoError(t, err)
		assert.NotEmpty(t, mismatch)
	})

	t.Run("too short", func(t *testing.T) {
		candidate := Tuple{
			int64(-8742),
			uint64(12342),
			"goodbye world",
		}
		pattern := Tuple{
			int64(-8742),
			Variable{},
			"goodbye world",
			Variable{},
		}

		mismatch, err := CompareTuples(pattern, candidate)
		assert.NoError(t, err)
		assert.NotEmpty(t, mismatch)
	})

	t.Run("maybe more", func(t *testing.T) {
		candidate := Tuple{
			int64(-8742),
			uint64(12342),
			"goodbye world",
			nil,
			Tuple{-55.1428},
		}
		pattern := Tuple{
			int64(-8742),
			Variable{},
			"goodbye world",
			MaybeMore{},
		}

		mismatch, err := CompareTuples(pattern, candidate)
		assert.NoError(t, err)
		assert.Empty(t, mismatch)
	})
}
