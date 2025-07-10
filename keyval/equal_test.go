package keyval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNil_Eq(t *testing.T) {
	x := Nil{}
	assert.True(t, x.Eq(Nil{}))
	assert.False(t, x.Eq(nil))
	assert.False(t, x.Eq(5))
}

func TestInt_Eq(t *testing.T) {
	x := Int(5)
	assert.True(t, x.Eq(Int(5)))
	assert.False(t, x.Eq(Int(6)))
	assert.False(t, x.Eq(Float(5)))
}

func TestUint_Eq(t *testing.T) {
	x := Uint(12)
	assert.True(t, x.Eq(Uint(12)))
	assert.False(t, x.Eq(Uint(2)))
	assert.False(t, x.Eq(Int(12)))
}

func TestBool_Eq(t *testing.T) {
	x := Bool(true)
	assert.True(t, x.Eq(Bool(true)))
	assert.False(t, x.Eq(Bool(false)))
	assert.False(t, x.Eq(Int(0)))
}

func TestFloat_Eq(t *testing.T) {
	x := Float(55.2)
	assert.True(t, x.Eq(Float(55.2)))
	assert.False(t, x.Eq(Float(22)))
	assert.False(t, x.Eq(Bool(true)))
}

// TODO: Add support for BigInt.
/*
func TestBigInt_Eq(t *testing.T) {
	x := BigInt(*big.NewInt(25))
	assert.True(t, x.Eq(BigInt(*big.NewInt(25))))
	assert.False(t, x.Eq(BigInt(*big.NewInt(60009))))
	assert.False(t, x.Eq(String("hi")))
}
*/

func TestString_Eq(t *testing.T) {
	x := String("hi world")
	assert.True(t, x.Eq(String("hi world")))
	assert.False(t, x.Eq(String("you")))
	assert.False(t, x.Eq(Int(-5)))
}

func TestUUID_Eq(t *testing.T) {
	x := UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}
	assert.True(t, x.Eq(UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}))
	assert.False(t, x.Eq(UUID{0x00, 0x00, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}))
	assert.False(t, x.Eq(String("wow")))
}

func TestBytes_Eq(t *testing.T) {
	x := Bytes{0xAB, 0xFF, 0x23}
	assert.True(t, x.Eq(Bytes{0xAB, 0xFF, 0x23}))
	assert.False(t, x.Eq(Bytes{0x00, 0xFF, 0x23}))
	assert.False(t, x.Eq(Uint(20)))
}

func TestVariable_Eq(t *testing.T) {
	x := Variable{IntType, StringType, UUIDType}
	assert.True(t, x.Eq(Variable{IntType, StringType, UUIDType}))
	assert.False(t, x.Eq(Variable{IntType, BoolType, UUIDType}))
	assert.False(t, x.Eq(Variable{IntType, StringType}))
	assert.False(t, x.Eq(Int(0)))
}

func TestMaybeMore_Eq(t *testing.T) {
	x := MaybeMore{}
	assert.True(t, x.Eq(MaybeMore{}))
	assert.False(t, x.Eq(Int(0)))
}

func TestTuple_Eq(t *testing.T) {
	x := Tuple{Int(1), Float(2.2), Tuple{Bool(true), Bool(false)}}
	assert.True(t, x.Eq(Tuple{Int(1), Float(2.2), Tuple{Bool(true), Bool(false)}}))
	assert.False(t, x.Eq(Tuple{Int(1), Float(2.2), Tuple{Bool(true), Int(0)}}))
	assert.False(t, x.Eq(Int(55)))
}

func TestClear_Eq(t *testing.T) {
	x := Clear{}
	assert.True(t, x.Eq(Clear{}))
	assert.False(t, x.Eq(Bool(false)))
}

func TestVStamp_Eq(t *testing.T) {
	x := VStamp{
		UserVersion: 587,
		TxVersion: [10]byte{
			0xa7, 0x32, 0x81, 0x89, 0xf3,
			0xfc, 0xf4, 0xa2, 0xb4, 0x99,
		},
	}
	assert.True(t, x.Eq(VStamp{
		UserVersion: 587,
		TxVersion: [10]byte{
			0xa7, 0x32, 0x81, 0x89, 0xf3,
			0xfc, 0xf4, 0xa2, 0xb4, 0x99,
		},
	}))
	assert.False(t, x.Eq(VStamp{
		UserVersion: 34,
		TxVersion: [10]byte{
			0xa7, 0x32, 0x81, 0x89, 0xf3,
			0xfc, 0xf4, 0xa2, 0xb4, 0x99,
		},
	}))
	assert.False(t, x.Eq(Float(22.8)))
}

func TestVStampFuture_Eq(t *testing.T) {
	x := VStampFuture{UserVersion: 641}
	assert.True(t, x.Eq(VStampFuture{UserVersion: 641}))
	assert.False(t, x.Eq(Int(33)))
}
