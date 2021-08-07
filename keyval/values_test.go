package keyval

import (
	"encoding/binary"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/stretchr/testify/assert"
)

var order = binary.BigEndian

func TestPackUnpackValue(t *testing.T) {
	tests := []struct {
		val interface{}
		typ ValueType
	}{
		{val: int64(-5548), typ: IntType},
		{val: uint64(128895), typ: UintType},
		{val: true, typ: BoolType},
		{val: float64(1288.9932), typ: FloatType},
		{val: "hola mundo", typ: StringType},
		{val: []byte{0xFF, 0xAA, 0xBC}, typ: BytesType},
		{val: tuple.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, typ: UUIDType},
		{val: Tuple{int64(225), float64(-55.8), "this is me"}, typ: TupleType},
	}

	for _, test := range tests {
		t.Run(string(test.typ), func(t *testing.T) {
			v, err := PackValue(order, test.val)
			assert.NoError(t, err)
			assert.NotNil(t, v)

			out, err := UnpackValue(order, test.typ, v)
			assert.NoError(t, err)
			assert.Equal(t, test.val, out)
		})
	}
}

func TestPackUnpackNil(t *testing.T) {
	v, err := PackValue(order, nil)
	assert.NoError(t, err)
	assert.Nil(t, v)

	out, err := UnpackValue(order, AnyType, v)
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestInvalidPackValue(t *testing.T) {
	out, err := PackValue(order, struct {
		f1 string
		f2 float32
	}{})
	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestInvalidUnpackValue(t *testing.T) {
	tests := []struct {
		val []byte
		typ ValueType
	}{
		{val: []byte{0x88, 0x10, 0xA2, 0xBB}, typ: IntType},
		{val: []byte{0x12, 0xA7, 0x0B}, typ: UintType},
		{val: []byte{0x12, 0xA7}, typ: BoolType},
		{val: []byte{0x88, 0x10, 0xA2, 0xBB, 0x74}, typ: FloatType},
		{val: []byte{0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81}, typ: UUIDType},
	}

	for _, test := range tests {
		t.Run(string(test.typ), func(t *testing.T) {
			out, err := UnpackValue(order, test.typ, test.val)
			assert.Error(t, err)
			assert.Nil(t, out)
		})
	}
}
