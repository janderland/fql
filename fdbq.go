package fdbq

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	Query struct {
		Key   *Key
		Value *Value
	}

	Key struct {
		Directory Directory
		Tuple     *Tuple
	}

	Directory []string

	Tuple []string

	Value string
)

func ParseQuery(query string) (*Query, error) {
	parts := strings.Split(query, "=")
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

func ParseKey(str string) (*Key, error) {
	parts := strings.SplitN(str, "(", 2)

	var directoryStr string
	var tupleStr string
	var err error

	if len(parts) == 1 {
		part := parts[0]
		if part[0] == '/' {
			directoryStr = part
		} else if part[0] == '(' {
			tupleStr = part
		} else {
			return nil, errors.New("key must start with either a directory '/' or a tuple '('")
		}
	} else {
		directoryStr = parts[0]
		tupleStr = parts[1]
	}

	directory, err := ParseDirectory(directoryStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse directory - %s", directoryStr)
	}
	tuple, err := ParseTuple(tupleStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse tuple - %s", tupleStr)
	}

	return &Key{
		Directory: directory,
		Tuple:     tuple,
	}, nil
}

func ParseDirectory(str string) (Directory, error) {
	first, str := str[0], str[1:]
	if first != '/' {
		return nil, errors.New("directory path must start with a '/'")
	}

	var directories []string
	var part string

	for len(str) > 0 {
		i := 0
		for i < len(str) && str[i] != '/' {
			i++
		}
		if i == 0 {
			return nil, errors.Errorf("%s part of directory path is empty", ordinal(len(directories)+1))
		}
		part, str = str[:i], str[i:]
		directories = append(directories, part)
	}

	return directories, nil
}

func ParseTuple(str string) (*Tuple, error) {
	return nil, nil
}

func ParseValue(str string) (*Value, error) {
	value := Value(str)
	return &value, nil
}

func ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return strconv.Itoa(x) + suffix
}
