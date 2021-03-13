package keyval

import (
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

func NormalizeTuple(tup Tuple) Tuple {
	for i, e := range tup {
		switch e.(type) {
		case int:
			tup[i] = int64(e.(int))
		case uint:
			tup[i] = uint64(e.(uint))
		case float32:
			tup[i] = float64(e.(float32))
		case tuple.Tuple:
			tup[i] = NormalizeTuple(FromFDBTuple(e.(tuple.Tuple)))
		case big.Int:
			val := e.(big.Int)
			tup[i] = &val
		}
	}
	return tup
}

func CompareTuples(pattern Tuple, candidate Tuple) ([]int, error) {
	if len(pattern) == 0 {
		return nil, errors.New("empty pattern")
	}

	switch pattern[len(pattern)-1].(type) {
	case MaybeMore:
		pattern = pattern[:len(pattern)-1]
	default:
		if len(pattern) < len(candidate) {
			return []int{len(pattern) + 1}, nil
		}
	}

	if len(pattern) > len(candidate) {
		return []int{len(candidate) + 1}, nil
	}

	var index []int
	err := ParseTuple(candidate, AllowLong, func(p *TupleParser) error {
		for i, e := range pattern {
			switch e.(type) {
			case int64:
				if p.Int() != e.(int64) {
					index = []int{i}
					return nil
				}

			case uint64:
				if p.Uint() != e.(uint64) {
					index = []int{i}
					return nil
				}

			case string:
				if p.String() != e.(string) {
					index = []int{i}
					return nil
				}

			case float64:
				if p.Float() != e.(float64) {
					index = []int{i}
					return nil
				}

			case bool:
				if p.Bool() != e.(bool) {
					index = []int{i}
					return nil
				}

			case nil:
				if e != p.Any() {
					index = []int{i}
					return nil
				}

			case *big.Int:
				if p.BigInt().Cmp(e.(*big.Int)) != 0 {
					index = []int{i}
					return nil
				}

			case Variable:
				break

			case Tuple:
				subIndex, err := CompareTuples(e.(Tuple), p.Tuple())
				if err != nil {
					return errors.Wrap(err, "failed to compare sub-tuple")
				}
				if len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}

			default:
				index = []int{i}
				return nil
			}
		}
		return nil
	})
	if err != nil {
		if c, ok := err.(ConversionError); ok {
			return []int{c.Index}, nil
		}
		return nil, errors.Wrap(err, "unexpected error")
	}
	return index, nil
}
