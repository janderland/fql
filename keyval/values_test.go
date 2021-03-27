package keyval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{val: UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, typ: UUIDType},
		{val: Tuple{int64(225), float64(-55.8), "this is me"}, typ: TupleType},
	}

	for _, test := range tests {
		t.Run(string(test.typ), func(t *testing.T) {
			v, err := PackValue(test.val)
			assert.NoError(t, err)
			assert.NotNil(t, v)

			out, err := UnpackValue(test.typ, v)
			assert.NoError(t, err)
			assert.Equal(t, test.val, out)
		})
	}
}
