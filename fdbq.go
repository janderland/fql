package fdbq

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	querySeparator = "="
)

type (
	Query struct {
		Key   *Key
		Value *Value
	}

	Key struct {
		Directory []string
		Tuple     []string
	}

	Value string
)

func ParseQuery(query string) (*Query, error) {
	parts := strings.Split(query, querySeparator)
	if len(parts) == 1 {
		return nil, errors.New("query missing '=' separator between key and value")
	} else if len(parts) > 2 {
		return nil, errors.New("query should only contain a single '='")
	}

	keyStr := parts[0]
	valueStr := parts[1]

	key, err := ParseKey(keyStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse key - %s", keyStr)
	}
	value, err := ParseValue(valueStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse value - %s", valueStr)
	}

	return &Query{
		Key:   key,
		Value: value,
	}, nil
}

func ParseKey(key string) (*Key, error) {

}

func ParseValue(value string) (*Value, error) {
	v := Value(value)
	return &v, nil
}
