package keyval

import (
	"math"
	"math/big"
	"testing"
)

func TestCompareTuples(t *testing.T) {
	Tuple := Tuple{
		int64(-8742),
		uint64(12342),
		"goodbye world",
		float64(-55.93),
		true,
		nil,
		big.NewInt(math.MaxInt64),
	}
}
