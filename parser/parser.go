package parser

import (
	"encoding/hex"
	"fmt"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"strconv"
	"strings"

	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

func ParseQuery(str string) (*keyval.KeyValue, error) {
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

	return &keyval.KeyValue{
		Key:   *key,
		Value: value,
	}, nil
}

func ParseKey(str string) (*keyval.Key, error) {
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

	key := &keyval.Key{}
	var err error
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

func ParseDirectory(str string) (keyval.Directory, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != '/' {
		return nil, errors.New("directory path must start with a '/'")
	}
	if str[len(str)-1] == '/' {
		return nil, errors.New("directory path shouldn't have a trailing '/'")
	}

	var directory keyval.Directory
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

func ParseTuple(str string) (keyval.Tuple, error) {
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
		return keyval.Tuple{}, nil
	}

	var tuple keyval.Tuple
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

const (
	Nil = "nil"
	True = "true"
	False = "false"

	VarStart = '{'
	VarSep = '|'
	VarEnd = '}'
)

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

func StringData(in interface{}) string {
	switch in := in.(type) {
	case nil:
		return Nil

	case bool:
		if in {
			return True
		} else {
			return False
		}

	case keyval.Variable:
		return StringVariable(in)

	default:
		var str strings.Builder
		str.WriteRune(VarStart)
		str.WriteString("invalid:")
		str.WriteString(fmt.Sprintf("%T(%v) ", in, in))
		str.WriteRune(VarEnd)
		return str.String()
	}
}

func ParseVariable(str string) (keyval.Variable, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != '{' {
		return nil, errors.New("variable must start with '{'")
	}
	if str[len(str)-1] != '}' {
		return nil, errors.New("variable must end with '}'")
	}

	var variable keyval.Variable
	if typeStr := str[1 : len(str)-1]; len(typeStr) > 0 {
	loop:
		for i, typeStr := range strings.Split(typeStr, string(VarSep)) {
			for _, v := range keyval.AllTypes() {
				if string(v) == typeStr {
					variable = append(variable, keyval.ValueType(typeStr))
					continue loop
				}
			}
			return nil, errors.Errorf("failed to parse %s type - %s", ordinal(i+1), typeStr)
		}
	}
	return variable, nil
}

func StringVariable(in keyval.Variable) string {
	var str strings.Builder
	str.WriteRune(VarStart)
	for i, typ := range in {
		str.WriteString(string(typ))
		if i != len(in)-1 {
			str.WriteRune(VarSep)
		}
	}
	str.WriteRune(VarEnd)
	return str.String()
}

func ParseString(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("input is empty")
	}
	if str[0] != '"' {
		return "", errors.New("strings must start with double quotes")
	}
	if str[len(str)-1] != '"' {
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

func ParseValue(str string) (keyval.Value, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == "clear" {
		return keyval.Clear{}, nil
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
