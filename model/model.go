package model

import "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

type (
	Query struct {
		Key   *Key
		Value *Value
	}

	Key struct {
		Directory []string
		Tuple     tuple.Tuple
	}

	Value string
)
