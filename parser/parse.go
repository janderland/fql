package parser

import (
	"encoding/hex"
	"fmt"
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
			return nil, false, err
		}
		return kv, false, nil
	}
	if strings.Contains(str, string(TupStart)) {
		key, err := ParseKey(str)
		if err != nil {
			return nil, false, err
		}
		return &q.KeyValue{Key: *key, Value: q.Variable{}}, false, nil
	}
	dir, err := ParseDirectory(str)
	if err != nil {
		return nil, false, err
	}
	return &q.KeyValue{Key: q.Key{Directory: dir}}, true, nil
}

func ParseKeyValue(str string) (*q.KeyValue, error) {
	wrapErr := newErrWrapper("key-value", str)
	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
	}

	parts := strings.Split(str, string(KVSep))
	if len(parts) == 1 {
		return nil, wrapErr(errors.New("missing '=' separator between key and value"))
	} else if len(parts) > 2 {
		return nil, wrapErr(errors.New("contains multiple '='"))
	}

	keyStr := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	key, err := ParseKey(keyStr)
	if err != nil {
		return nil, wrapErr(err)
	}
	value, err := ParseValue(valueStr)
	if err != nil {
		return nil, wrapErr(err)
	}

	return &q.KeyValue{
		Key:   *key,
		Value: value,
	}, nil
}

func ParseKey(str string) (*q.Key, error) {
	wrapErr := newErrWrapper("key", str)

	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
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
			return nil, wrapErr(err)
		}
	}
	if len(tupleStr) > 0 {
		key.Tuple, err = ParseTuple(tupleStr)
		if err != nil {
			return nil, wrapErr(err)
		}
	}
	return key, nil
}

func ParseDirectory(str string) (q.Directory, error) {
	wrapErr := newErrWrapper("directory", str)

	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
	}
	if str[0] != DirSep {
		return nil, wrapErr(errors.New("directory path must start with a '/'"))
	}
	if str[len(str)-1] == '/' {
		return nil, wrapErr(errors.New("directory path shouldn't have a trailing '/'"))
	}

	var directory q.Directory
	for i, part := range strings.Split(str[1:], "/") {
		element, err := ParsePathElement(i, part)
		if err != nil {
			return nil, wrapErr(err)
		}
		directory = append(directory, element)
	}
	return directory, nil
}

func ParsePathElement(i int, str string) (interface{}, error) {
	wrapErr := newErrWrapper(fmt.Sprintf("%s element", ordinal(i+1)), str)

	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return nil, wrapErr(errors.New("path element is empty"))
	}
	if str[0] == '{' {
		variable, err := ParseVariable(str)
		if err != nil {
			return nil, wrapErr(err)
		}
		return variable, nil
	} else {
		return str, nil
	}
}

func ParseTuple(str string) (q.Tuple, error) {
	wrapErr := newErrWrapper("tuple", str)

	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
	}

	if str[0] != '(' {
		return nil, wrapErr(errors.New("tuple must start with a '('"))
	}
	if str[len(str)-1] != ')' {
		return nil, wrapErr(errors.New("tuple must end with a ')"))
	}

	tup := q.Tuple{}

	str = str[1 : len(str)-1]
	if len(str) == 0 {
		return tup, nil
	}

	for i, elementStr := range strings.Split(str, ",") {
		element, err := ParseTupleElement(i, elementStr)
		if err != nil {
			return nil, wrapErr(err)
		}
		tup = append(tup, element)
	}
	return tup, nil
}

func ParseTupleElement(i int, str string) (interface{}, error) {
	wrapErr := newErrWrapper(fmt.Sprintf("%s element", ordinal(i+1)), str)

	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return nil, wrapErr(errors.New("element is empty"))
	}

	var element interface{}
	var err error
	if str[0] == '(' {
		element, err = ParseTuple(str)
	} else {
		element, err = ParseData(str)
	}
	if err != nil {
		return nil, wrapErr(err)
	}
	return element, nil
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
		return data, err
	}
	if str[0] == '"' {
		data, err := ParseString(str)
		return data, err
	}
	if strings.Count(str, "-") == 4 {
		data, err := ParseUUID(str)
		return data, err
	}
	data, err := ParseNumber(str)
	return data, err
}

func ParseVariable(str string) (q.Variable, error) {
	wrapErr := newErrWrapper("variable", str)

	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
	}
	if str[0] != '{' {
		return nil, wrapErr(errors.New("variable must start with '{'"))
	}
	if str[len(str)-1] != '}' {
		return nil, wrapErr(errors.New("variable must end with '}'"))
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

			wrapTypeErr := newErrWrapper(fmt.Sprintf("%s type", ordinal(i+1)), typeStr)
			return nil, wrapErr(wrapTypeErr(errors.Errorf("invalid type")))
		}
	}
	return variable, nil
}

func ParseString(str string) (string, error) {
	wrapErr := newErrWrapper("string", str)
	if len(str) == 0 {
		return "", wrapErr(errors.New("input is empty"))
	}
	if str[0] != StrStart {
		return "", wrapErr(errors.New("strings must start with double quotes"))
	}
	if str[len(str)-1] != StrEnd {
		return "", wrapErr(errors.New("strings must end with double quotes"))
	}
	return str[1 : len(str)-1], nil
}

func ParseUUID(str string) (tuple.UUID, error) {
	wrapErr := newErrWrapper("UUID", str)

	if len(str) == 0 {
		return tuple.UUID{}, wrapErr(errors.New("input is empty"))
	}

	groups := strings.Split(str, "-")
	checkLen := func(i int, expLen int) error {
		if len(groups[i]) != expLen {
			err := errors.Errorf("the %s group should contain %d characters rather than %d", ordinal(i+1), expLen, len(groups[i]))
			return wrapErr(err)
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
		return tuple.UUID{}, wrapErr(err)
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
	wrapErr := newErrWrapper("number", str)
	return nil, wrapErr(errors.New("invalid syntax"))
}

func ParseValue(str string) (q.Value, error) {
	wrapErr := newErrWrapper("value", str)
	if len(str) == 0 {
		return nil, wrapErr(errors.New("input is empty"))
	}
	if str == Clear {
		return q.Clear{}, nil
	}
	if str[0] == '(' {
		out, err := ParseTuple(str)
		return out, wrapErr(err)
	}
	out, err := ParseData(str)
	return out, wrapErr(err)
}
