package values

import (
	"encoding/binary"
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

var order = binary.BigEndian

func TestPackUnpack(t *testing.T) {
	tests := []struct {
		val q.Value
		typ q.ValueType
	}{
		{val: q.Int(-5548), typ: q.IntType},
		{val: q.Uint(128895), typ: q.UintType},
		{val: q.Bool(true), typ: q.BoolType},
		{val: q.Float(1288.9932), typ: q.FloatType},
		{val: q.String("hola mundo"), typ: q.StringType},
		{val: q.Bytes{0xFF, 0xAA, 0xBC}, typ: q.BytesType},
		{val: q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, typ: q.UUIDType},
		{val: q.Tuple{q.Int(225), q.Float(-55.8), q.String("this is me")}, typ: q.TupleType},
	}

	for _, test := range tests {
		t.Run(string(test.typ), func(t *testing.T) {
			v, err := Pack(test.val, order)
			assert.NoError(t, err)
			assert.NotNil(t, v)

			out, err := Unpack(v, test.typ, order)
			assert.NoError(t, err)
			assert.Equal(t, test.val, out)
		})
	}
}

func TestPackUnpackNil(t *testing.T) {
	v, err := Pack(nil, order)
	assert.Error(t, err)
	assert.Nil(t, v)

	out, err := Unpack(nil, q.AnyType, order)
	assert.NoError(t, err)
	assert.Equal(t, q.Bytes(nil), out)
}

func TestInvalidUnpack(t *testing.T) {
	tests := []struct {
		val []byte
		typ q.ValueType
	}{
		{val: []byte{0x88, 0x10, 0xA2, 0xBB}, typ: q.IntType},
		{val: []byte{0x12, 0xA7, 0x0B}, typ: q.UintType},
		{val: []byte{0x12, 0xA7}, typ: q.BoolType},
		{val: []byte{0x88, 0x10, 0xA2, 0xBB, 0x74}, typ: q.FloatType},
		{val: []byte{0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81}, typ: q.UUIDType},
	}

	for _, test := range tests {
		t.Run(string(test.typ), func(t *testing.T) {
			out, err := Unpack(test.val, test.typ, order)
			assert.Error(t, err)
			assert.Nil(t, out)
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name  string
		query q.Value
		val   []byte
		out   q.Value
	}{
		{name: "empty variable", query: q.Variable{}, val: []byte{0xAE, 0xBC}, out: q.Bytes{0xAE, 0xBC}},
		{name: "variable match", query: q.Variable{q.IntType, q.StringType}, val: []byte("hi"), out: q.String("hi")},
		{name: "variable mismatch", query: q.Variable{q.IntType}, val: []byte("hi"), out: nil},
		{name: "packed match", query: q.String("you"), val: []byte("you"), out: q.String("you")},
		{name: "packed mismatch", query: q.Int(22), val: []byte("you"), out: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter, err := NewDeserializer(test.query, binary.BigEndian, true)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			val, err := filter(test.val)
			assert.Equal(t, test.out, val)
			assert.NoError(t, err)
		})
	}
}
