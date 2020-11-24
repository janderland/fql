package model

type (
	Query struct {
		Key   *Key
		Value *Value
	}

	Key struct {
		Directory []string
		Tuple     Tuple
	}

	Tuple []interface{}

	Value string

	Variable struct{}

	Clear struct{}
)
