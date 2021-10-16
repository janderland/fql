// Package keyval provides types representing key-values and utilities
// for their inspection and manipulation. These types serve as both the
// AST which the FDBQ parser outputs and as part of the Go API for
// performing queries.
//
// Embedded Query Strings
//
// When working with SQL, programmers will often embed SQL strings in
// the application. This requires extra tooling to catch syntax errors
// at build time. Instead of using string literals, FDBQ allows the
// programmer to directly construct the queries using the type-safe AST,
// allowing some syntax errors to be caught at build time.
//
// This package does not prevent all kinds of syntax errors, only
// "structural" ones. For instance, the type system ensures tuples only
// contain valid elements and limits what kinds of objects can be used
// as the value of a key-value. In spite of this, an invalid query
// can still be constructed (see package "class").
//
// Visitor Pattern
//
// The Directory, Tuple, & Value types would be best represented by tagged
// unions. While Go does not natively support tagged unions, this package
// implements equivalent functionality using interfaces. Furthermore, these
// interfaces implement the visitor pattern which avoids the need for type
// switches. While type switches will almost certainly result in a faster
// runtime, the programmer must remember to handle all the types implementing
// a given interface. The visitor pattern forces the programmer to handle all
// relevant types via the type system.
package keyval

import "math/big"

//go:generate go run ./operation -op-name Query     -param-name query      -types Directory,Key,KeyValue
//go:generate go run ./operation -op-name Directory -param-name DirElement -types String,Variable
//go:generate go run ./operation -op-name Tuple     -param-name TupElement -types Tuple,Nil,Int,Uint,Bool,Float,BigInt,String,UUID,Bytes,Variable,MaybeMore
//go:generate go run ./operation -op-name Value     -param-name value      -types Tuple,Nil,Int,Uint,Bool,Float,String,UUID,Bytes,Variable,Clear

type (
	Query = query

	// A KeyValue is a query or result depending on the
	// context. If the KeyValue is a result, it will not
	// contain a Variable.
	KeyValue struct {
		Key   Key
		Value Value
	}

	// A Key represents an FDB key made up of a Directory
	// and optionally a Tuple. A Key cannot have both an
	// empty Directory and an empty Tuple.
	Key struct {
		Directory Directory
		Tuple     Tuple
	}

	// A Directory is equivalent to a path used by the
	// directory layer of the FDB API. A Directory may
	// contain instances of string or Variable.
	Directory []DirElement

	// A Tuple is similar to a tuple.Tuple. It may contain
	// anything in a valid tuple.Tuple in addition to
	// Variable and Tuple.
	Tuple []TupElement

	// A Value represents an FDB value stored alongside
	// a key. This type may contain nil, Int, Uint, Bool,
	// Float, BigInt, String, UUID, Bytes, Tuple, Variable,
	// or Clear.
	Value = value

	// A Variable is used as a placeholder for any valid
	// values within a type constraint.
	Variable []ValueType

	// A MaybeMore is a special kind of TupElement. It
	// may only appear as the last element of the Tuple.
	// It designates that the Tuple will match all Tuples
	// which contain a matching prefix.
	MaybeMore struct{}

	// A Clear is a special kind of Value which designates
	// a KeyValue as a clear operation.
	Clear struct{}

	Nil    struct{}
	Int    int64
	Uint   uint64
	Bool   bool
	Float  float64
	BigInt big.Int
	String string
	UUID   [16]byte
	Bytes  []byte
)

// ValueType specifies the variable's expected type.
type ValueType string

const (
	AnyType    ValueType = ""
	IntType    ValueType = "int"
	UintType   ValueType = "uint"
	BoolType   ValueType = "bool"
	FloatType  ValueType = "float"
	BigIntType ValueType = "bigint"
	StringType ValueType = "string"
	BytesType  ValueType = "bytes"
	UUIDType   ValueType = "uuid"
	TupleType  ValueType = "tuple"
)

func AllTypes() []ValueType {
	return []ValueType{
		AnyType,
		IntType,
		UintType,
		BoolType,
		FloatType,
		BigIntType,
		StringType,
		BytesType,
		UUIDType,
		TupleType,
	}
}
