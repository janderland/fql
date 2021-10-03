package compare

import (
	"math"
	"math/big"
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

func TestCompareTuples(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		candidate := q.Tuple{
			q.Int(-8742),
			q.Uint(12342),
			q.String("goodbye world"),
			q.Float(-55.93),
			q.Bool(true),
			q.Nil{},
			q.BigInt(*big.NewInt(math.MaxInt64)),
		}
		pattern := make(q.Tuple, len(candidate))
		assert.Equal(t, copy(pattern, candidate), len(candidate))

		mismatch := Tuples(pattern, candidate)
		assert.Empty(t, mismatch)
	})

	t.Run("variable", func(t *testing.T) {
		candidate1 := q.Tuple{
			q.Int(-8742),
			q.Uint(12342),
			q.String("goodbye world"),
			q.Nil{},
		}
		candidate2 := q.Tuple{
			q.Int(-8742),
			q.Tuple{q.String("nowhere"), q.BigInt(*big.NewInt(55))},
			q.String("goodbye world"),
			q.Int(555),
		}
		pattern := q.Tuple{
			q.Int(-8742),
			q.Variable{},
			q.String("goodbye world"),
			q.Variable{},
		}

		mismatch := Tuples(pattern, candidate1)
		assert.Empty(t, mismatch)

		mismatch = Tuples(pattern, candidate2)
		assert.Empty(t, mismatch)
	})

	t.Run("multi type", func(t *testing.T) {
		candidate := q.Tuple{q.String("where am i?")}
		pattern := q.Tuple{q.Variable{q.IntType, q.TupleType, q.StringType}}

		mismatch := Tuples(pattern, candidate)
		assert.Empty(t, mismatch)
	})

	t.Run("too long", func(t *testing.T) {
		candidate := q.Tuple{
			q.Int(-8742),
			q.Uint(12342),
			q.String("goodbye world"),
			q.Nil{},
			q.Tuple{q.Float(-55.1428)},
		}
		pattern := q.Tuple{
			q.Int(-8742),
			q.Variable{},
			q.String("goodbye world"),
			q.Variable{},
		}

		mismatch := Tuples(pattern, candidate)
		assert.NotEmpty(t, mismatch)
	})

	t.Run("too short", func(t *testing.T) {
		candidate := q.Tuple{
			q.Int(-8742),
			q.Uint(12342),
			q.String("goodbye world"),
		}
		pattern := q.Tuple{
			q.Int(-8742),
			q.Variable{},
			q.String("goodbye world"),
			q.Variable{},
		}

		mismatch := Tuples(pattern, candidate)
		assert.NotEmpty(t, mismatch)
	})

	t.Run("maybe more", func(t *testing.T) {
		candidate := q.Tuple{
			q.Int(-8742),
			q.Uint(12342),
			q.String("goodbye world"),
			q.Nil{},
			q.Tuple{q.Float(-55.1428)},
		}
		pattern := q.Tuple{
			q.Int(-8742),
			q.Variable{},
			q.String("goodbye world"),
			q.MaybeMore{},
		}

		mismatch := Tuples(pattern, candidate)
		assert.Empty(t, mismatch)
	})
}
