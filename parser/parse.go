package parser

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

func ParseQuery(str string) (*q.KeyValue, bool, error) {
	if strings.Contains(str, string(KVSep)) {
		kv, err := ParseKeyValue(str)
		if err != nil {
			return nil, false, errors.Wrap(err, "failed to parse as key-value")
		}
		return kv, false, nil
	}
	if strings.Contains(str, string(TupStart)) {
		key, err := ParseKey(str)
		if err != nil {
			return nil, false, errors.Wrap(err, "failed to parse as key")
		}
		return &q.KeyValue{Key: *key, Value: q.Variable{}}, false, nil
	}
	dir, err := ParseDirectory(str)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to parse as directory")
	}
	return &q.KeyValue{Key: q.Key{Directory: dir}}, true, nil
}

func ParseKeyValue(str string) (*q.KeyValue, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	parts := strings.Split(str, string(KVSep))
	if len(parts) == 1 {
		return nil, errors.New("query missing '=' separator between key and value")
	} else if len(parts) > 2 {
		return nil, errors.New("query should only contain a single '='")
	}

	keyStr := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	key, err := ParseKey(keyStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key")
	}
	value, err := ParseValue(valueStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse value")
	}

	return &q.KeyValue{
		Key:   *key,
		Value: value,
	}, nil
}

func ParseKey(str string) (*q.Key, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	var parts []string
	if i := strings.Index(str, "("); i == -1 {
		parts = []string{str, ""}
	} else {
		parts = []string{str[:i], str[i:]}
	}

	directoryStr := strings.TrimSpace(parts[0])
	tupleStr := strings.TrimSpace(parts[1])

	key := &q.Key{}
	var err error
	if len(directoryStr) > 0 {
		key.Directory, err = ParseDirectory(directoryStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse directory")
		}
	}
	if len(tupleStr) > 0 {
		key.Tuple, err = ParseTuple(tupleStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse tuple")
		}
	}
	return key, nil
}

func ParseDirectory(str string) (q.Directory, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != DirSep {
		return nil, errors.New("directory path must start with a '/'")
	}
	if str[len(str)-1] == '/' {
		return nil, errors.New("directory path shouldn't have a trailing '/'")
	}

	var directory q.Directory
	for i, part := range strings.Split(str[1:], "/") {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			return nil, errors.Errorf("%s part of directory path is empty", ordinal(i+1))
		}
		if part[0] == '{' {
			variable, err := ParseVariable(part)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse %s part of directory path as a variable - %s", ordinal(i+1), part)
			}
			directory = append(directory, variable)
		} else {
			directory = append(directory, part)
		}
	}
	return directory, nil
}

func ParseTuple(str string) (q.Tuple, error) {
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
		return q.Tuple{}, nil
	}

	var tup q.Tuple
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
		tup = append(tup, element)
	}
	return tup, nil
}

func ParseData(str string) (interface{}, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == Nil {
		return nil, nil
	}
	if str == True {
		return true, nil
	}
	if str == False {
		return false, nil
	}
	if str[0] == VarStart {
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

func ParseVariable(str string) (q.Variable, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != '{' {
		return nil, errors.New("variable must start with '{'")
	}
	if str[len(str)-1] != '}' {
		return nil, errors.New("variable must end with '}'")
	}

	var variable q.Variable
	if typeStr := str[1 : len(str)-1]; len(typeStr) > 0 {
	loop:
		for i, typeStr := range strings.Split(typeStr, string(VarSep)) {
			typeStr = strings.TrimSpace(typeStr)
			for _, v := range q.AllTypes() {
				if string(v) == typeStr {
					variable = append(variable, q.ValueType(typeStr))
					continue loop
				}
			}
			return nil, errors.Errorf("failed to parse %s type - %s", ordinal(i+1), typeStr)
		}
	}
	return variable, nil
}

func ParseString(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("input is empty")
	}
	if str[0] != StrStart {
		return "", errors.New("strings must start with double quotes")
	}
	if str[len(str)-1] != StrEnd {
		return "", errors.New("strings must end with double quotes")
	}
	return str[1 : len(str)-1], nil
}

func ParseUUID(str string) (tuple.UUID, error) {
	if len(str) == 0 {
		return tuple.UUID{}, errors.New("input is empty")
	}

	groups := strings.Split(str, "-")
	checkLen := func(i int, expLen int) error {
		if len(groups[i]) != expLen {
			return errors.Errorf("the %s group should contain %d characters rather than %d", ordinal(i+1), expLen, len(groups[i]))
		}
		return nil
	}
	if err := checkLen(0, 8); err != nil {
		return tuple.UUID{}, err
	}
	if err := checkLen(1, 4); err != nil {
		return tuple.UUID{}, err
	}
	if err := checkLen(2, 4); err != nil {
		return tuple.UUID{}, err
	}
	if err := checkLen(3, 4); err != nil {
		return tuple.UUID{}, err
	}
	if err := checkLen(4, 12); err != nil {
		return tuple.UUID{}, err
	}

	var uuid tuple.UUID
	_, err := hex.Decode(uuid[:], []byte(strings.ReplaceAll(str, "-", "")))
	if err != nil {
		return tuple.UUID{}, errors.Wrap(err, "failed to decode hexadecimal string")
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

func ParseValue(in string) (q.Value, error) {
	if len(in) == 0 {
		return nil, errors.New("input is empty")
	}
	if in == Clear {
		return q.Clear{}, nil
	}
	if in[0] == '(' {
		out, err := ParseTuple(in)
		return out, errors.Wrap(err, "failed to parse as tuple")
	}
	out, err := ParseData(in)
	return out, errors.Wrap(err, "failed to parse as data")
}
