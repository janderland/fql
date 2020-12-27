package parser

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/janderland/fdbq/kv"
	"github.com/pkg/errors"
)

func ParseQuery(str string) (*kv.KeyValue, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	parts := strings.Split(str, "=")
	if len(parts) == 1 {
		return nil, errors.New("query missing '=' separator between key and value")
	} else if len(parts) > 2 {
		return nil, errors.New("query should only contain a single '='")
	}

	keyStr := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	key, err := ParseKey(keyStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse key - %s", keyStr)
	}
	value, err := ParseValue(valueStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse value - %s", valueStr)
	}

	return &kv.KeyValue{
		Key:   *key,
		Value: value,
	}, nil
}

func ParseKey(str string) (*kv.Key, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	var parts []string
	if i := strings.Index(str, "("); i == -1 {
		parts = []string{str}
	} else {
		parts = []string{str[:i], str[i:]}
	}

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

	directoryStr = strings.TrimSpace(directoryStr)
	tupleStr = strings.TrimSpace(tupleStr)

	key := &kv.Key{}
	if len(directoryStr) > 0 {
		key.Directory, err = ParseDirectory(directoryStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse directory - %s", directoryStr)
		}
	}
	if len(tupleStr) > 0 {
		key.Tuple, err = ParseTuple(tupleStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse tuple - %s", tupleStr)
		}
	}
	return key, nil
}

func ParseDirectory(str string) (kv.Directory, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != '/' {
		return nil, errors.New("directory path must start with a '/'")
	}
	if str[len(str)-1] == '/' {
		return nil, errors.New("directory path shouldn't have a trailing '/'")
	}

	var directory kv.Directory
	for i, part := range strings.Split(str[1:], "/") {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			return nil, errors.Errorf("%s part of directory path is empty", ordinal(i+1))
		}
		if part[0] == '{' {
			variable, err := ParseVariable(part)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse %s part of directory path as a variable", ordinal(i+1))
			}
			directory = append(directory, *variable)
		} else {
			directory = append(directory, part)
		}
	}
	return directory, nil
}

func ParseTuple(str string) (kv.Tuple, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	if str[0] != '(' {
		return nil, errors.New("tuple must start with a '('")
	}
	if str[len(str)-1] != ')' {
		return nil, errors.New("tuple must end with a ')")
	}

	str = str[1 : len(str)-1]
	if len(str) == 0 {
		return kv.Tuple{}, nil
	}

	var tuple kv.Tuple
	for i, elementStr := range strings.Split(str, ",") {
		var element interface{}
		var err error

		elementStr = strings.TrimSpace(elementStr)
		if len(elementStr) == 0 {
			return nil, errors.Errorf("%s element is empty", ordinal(i+1))
		}

		if elementStr[0] == '(' {
			element, err = ParseTuple(elementStr)
		} else {
			element, err = ParseData(elementStr)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s element - %s", ordinal(i+1), elementStr)
		}
		tuple = append(tuple, element)
	}
	return tuple, nil
}

func ParseData(str string) (interface{}, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == "nil" {
		return nil, nil
	}
	if str == "true" {
		return true, nil
	}
	if str == "false" {
		return false, nil
	}
	if str[0] == '{' {
		data, err := ParseVariable(str)
		return data, errors.Wrap(err, "failed to parse as variable")
	}
	if str[0] == '"' {
		data, err := ParseString(str)
		return data, errors.Wrap(err, "failed to parse as string")
	}
	if strings.Count(str, "-") == 4 {
		data, err := ParseUUID(str)
		return data, errors.Wrap(err, "failed to parse as UUID")
	}
	data, err := ParseNumber(str)
	return data, errors.Wrap(err, "failed to parse as number")
}

func ParseVariable(str string) (*kv.Variable, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != '{' {
		return nil, errors.New("variable must start with '{'")
	}
	if str[len(str)-1] != '}' {
		return nil, errors.New("variable must end with '}'")
	}
	return &kv.Variable{Name: str[1 : len(str)-1]}, nil
}

func ParseString(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("input is empty")
	}
	if str[0] != '"' {
		return "", errors.New("strings must start with single quotes")
	}
	if str[len(str)-1] != '"' {
		return "", errors.New("strings must end with single quotes")
	}
	return str[1 : len(str)-1], nil
}

func ParseUUID(str string) (kv.UUID, error) {
	if len(str) == 0 {
		return kv.UUID{}, errors.New("input is empty")
	}

	groups := strings.Split(str, "-")
	checkLen := func(i int, expLen int) error {
		if len(groups[i]) != expLen {
			return errors.Errorf("the %s group should contain %d characters rather than %d", ordinal(i+1), expLen, len(groups[i]))
		}
		return nil
	}
	if err := checkLen(0, 8); err != nil {
		return kv.UUID{}, err
	}
	if err := checkLen(1, 4); err != nil {
		return kv.UUID{}, err
	}
	if err := checkLen(2, 4); err != nil {
		return kv.UUID{}, err
	}
	if err := checkLen(3, 4); err != nil {
		return kv.UUID{}, err
	}
	if err := checkLen(4, 12); err != nil {
		return kv.UUID{}, err
	}

	var uuid kv.UUID
	n, err := hex.Decode(uuid[:], []byte(strings.ReplaceAll(str, "-", "")))
	if err != nil {
		return kv.UUID{}, errors.Wrap(err, "failed to decode hexadecimal string")
	}
	if n != 16 {
		return kv.UUID{}, errors.Errorf("decoded %d bytes instead of 16", n)
	}
	return uuid, nil
}

func ParseNumber(str string) (interface{}, error) {
	i, iErr := strconv.ParseInt(str, 10, 64)
	if iErr == nil {
		return i, nil
	}
	u, uErr := strconv.ParseUint(str, 10, 64)
	if uErr == nil {
		return u, nil
	}
	f, fErr := strconv.ParseFloat(str, 64)
	if fErr == nil {
		return f, nil
	}
	return nil, errors.Errorf("%v, %v, %v", iErr.Error(), uErr.Error(), fErr.Error())
}

func ParseValue(str string) (kv.Value, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == "clear" {
		return kv.Clear{}, nil
	}
	if str[0] == '(' {
		return ParseTuple(str)
	}
	return ParseData(str)
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
