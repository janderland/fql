// Package keyval provides types representing key-values and utilities
// for their inspection and manipulation. These types model both queries
// and the data returned by queries. These types can be constructed from
// query strings by the Parser but are also designed to be easily
// constructed directly in a Go app.
//
// Embedded Query Strings
//
// When working with SQL, programmers will often embed SQL strings in
// the application. This requires extra tooling to catch syntax errors
// at build time. Instead of using string literals, FDBQ allows the
// programmer to directly construct the queries using the types in this
// package, allowing some syntax errors to be caught at build time.
//
// This package does not prevent all kinds of syntax errors, only
// "structural" ones. For instance, the type system ensures tuples only
// contain valid elements and limits what kinds of objects can be used
// as the value of a key-value. In spite of this, invalid queries can
// still be constructed (see package "class").
//
// Operations (Visitor Pattern)
//
// The Directory, Tuple, & Value types would be best represented by tagged
// unions. While Go does not natively support tagged unions, this package
// implements equivalent functionality using the visitor pattern.
//
// Instead of implementing a type switch which handles every type in the
// union, a visitor interface is implemented with methods handling each
// type. The correct method is called at runtime via some generated glue
// code. This glue code also defines an interface for the union itself,
// allowing us to avoid using interface{}.
//
// Structs implementing these visitor interfaces define a parameterized
// (generic) function for the types in the union. For this reason, they
// are called "operations" rather than "visitors" in this codebase (see
// package "operation").
//
// Data Types
//
// There are a special group of types defined in this package named the
// "primitive" types. These include Nil, Int, Uint, Bool, Float, String,
// UUID, and Bytes. All of these types can be used as a TupElement or as
// a Value. When used as a TupElement, they are serialized by FDB tuple
// packing. When used as a Value, they are known as a "primitive" value
// and are serialized by FDBQ.
package keyval

import "math/big"

//go:generate go run ./operation -op-name Query     -param-name query      -types Directory,Key,KeyValue
//go:generate go run ./operation -op-name Directory -param-name DirElement -types String,Variable
//go:generate go run ./operation -op-name Tuple     -param-name TupElement -types Tuple,Nil,Int,Uint,Bool,Float,BigInt,String,UUID,Bytes,Variable,MaybeMore
//go:generate go run ./operation -op-name Value     -param-name value      -types Tuple,Nil,Int,Uint,Bool,Float,String,UUID,Bytes,Variable,Clear

type (
	// Query is an interface implemented by the types which can
	// be passed to engine.Engine as a query. This includes
	// KeyValue, Key, & Directory.
	Query = query

	// KeyValue can be passed to engine.Engine as a query
	// or be returned from engine.Engine as a query's result.
	// When returned as a result, KeyValue will not contain a
	// Variable, Clear, or MaybeMore.
	KeyValue struct {
		Key   Key
		Value Value
	}

	// Key can be passed to engine.Engine as a query or be
	// returned from engine.Engine as part of a query's result.
	// When used as a query, it is equivalent to a KeyValue
	// whose Value is an empty Variable. When returned as a
	// result, it will not contain a Variable or MaybeMore.
	Key struct {
		Directory Directory
		Tuple     Tuple
	}

	// Directory can be passed to engine.Engine as a query or
	// be returned from engine.Engine as part of a query's
	// result. When used as a query, only the directory layer
	// is accessed. It may contain String or Variable.
	Directory []DirElement

	// Tuple may contain another Tuple, Variable, MaybeMore,
	// or any of the "primitive" types.
	Tuple []TupElement

	// Value may contain Tuple, Variable, Clear, or any
	// of the "primitive" types.
	Value = value

	// Variable
	// TODO: Documentation
	Variable []ValueType

	// MaybeMore is a special kind of TupElement. It may only
	// appear as the last element of the Tuple. It causes a
	// query to access all key-values whose keys are prefixed
	// by the query's key.
	MaybeMore struct{}

	// Clear is a special kind of Value which designates
	// a KeyValue as a clear operation.
	Clear struct{}
)

// These are the "primitive" types.
type (
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
