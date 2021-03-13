package keyval

import (
	"fmt"
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

type TupleFlag = int

const (
	AllErrors TupleFlag = iota
	AllowLong
)

type ConversionError struct {
	InValue interface{}
	OutType interface{}
	Index   int
}

func (t ConversionError) Error() string {
	return fmt.Sprintf("failed to convert element %d from %v to %T",
		t.Index, t.InValue, t.OutType)
}

var (
	ShortTupleError = errors.New("read past end of tuple")
	LongTupleError  = errors.New("did not parse entire tuple")
)

type TupleParser struct {
	t Tuple
	i int
}

func ParseTuple(t Tuple, flag TupleFlag, f func(p *TupleParser) error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if e, ok := e.(ConversionError); ok {
				err = e
				return
			}
			if e == ShortTupleError {
				err = ShortTupleError
				return
			}
			panic(e)
		}
	}()

	p := TupleParser{t: t}
	if err := f(&p); err != nil {
		return err
	}

	if flag == AllErrors && p.i != len(t) {
		return LongTupleError
	}
	return nil
}

func (p *TupleParser) getIndex() int {
	if p.i >= len(p.t) {
		panic(ShortTupleError)
	}

	p.i++
	return p.i - 1
}

func (p *TupleParser) Any() interface{} {
	return p.t[p.getIndex()]
}

func (p *TupleParser) Bool() (out bool) {
	i := p.getIndex()
	if val, ok := p.t[i].(bool); ok {
		return val
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) String() (out string) {
	i := p.getIndex()
	if val, ok := p.t[i].(string); ok {
		return val
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) Int() (out int64) {
	i := p.getIndex()
	if val, ok := p.t[i].(int64); ok {
		return val
	}
	if val, ok := p.t[i].(int); ok {
		return int64(val)
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) Uint() (out uint64) {
	i := p.getIndex()
	if val, ok := p.t[i].(int64); ok {
		return uint64(val)
	}
	if val, ok := p.t[i].(uint64); ok {
		return val
	}
	if val, ok := p.t[i].(int); ok {
		return uint64(val)
	}
	if val, ok := p.t[i].(uint); ok {
		return uint64(val)
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) BigInt() (out *big.Int) {
	i := p.getIndex()
	if val, ok := p.t[i].(int64); ok {
		return big.NewInt(val)
	}
	if val, ok := p.t[i].(int); ok {
		return big.NewInt(int64(val))
	}
	if val, ok := p.t[i].(uint64); ok {
		out = big.NewInt(0)
		out.SetUint64(val)
		return out
	}
	if val, ok := p.t[i].(uint); ok {
		out = big.NewInt(0)
		out.SetUint64(uint64(val))
		return out
	}
	if val, ok := p.t[i].(big.Int); ok {
		return &val
	}
	if val, ok := p.t[i].(*big.Int); ok {
		return val
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) Float() (out float64) {
	i := p.getIndex()
	if val, ok := p.t[i].(float64); ok {
		return val
	}
	if val, ok := p.t[i].(float32); ok {
		return float64(val)
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) UUID() (out UUID) {
	i := p.getIndex()
	if val, ok := p.t[i].(UUID); ok {
		return val
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) Bytes() (out []byte) {
	i := p.getIndex()
	if val, ok := p.t[i].([]byte); ok {
		return val
	}
	if val, ok := p.t[i].(fdb.KeyConvertible); ok {
		return val.FDBKey()
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}

func (p *TupleParser) Tuple() (out Tuple) {
	i := p.getIndex()
	if val, ok := p.t[i].(Tuple); ok {
		return val
	}
	if val, ok := p.t[i].(tuple.Tuple); ok {
		return FromFDBTuple(val)
	}
	panic(ConversionError{
		InValue: p.t[i],
		OutType: out,
		Index:   i,
	})
}
