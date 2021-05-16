package parser

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

	"github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

const (
	KVSep  = '='
	DirSep = '/'

	TupStart = '('
	TupEnd   = ')'
	TupSep   = ','

	VarStart = '{'
	VarEnd   = '}'
	VarSep   = '|'

	StrStart = '"'
	StrEnd   = '"'

	Nil   = "nil"
	True  = "true"
	False = "false"
	Clear = "clear"
)

func ParseKeyValue(str string) (*keyval.KeyValue, error) {
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

func FormatKeyValue(kv keyval.KeyValue) (string, error) {
	key, err := FormatKey(kv.Key)
	if err != nil {
		return "", errors.Wrap(err, "failed to format key")
	}
	val, err := FormatValue(kv.Value)
	if err != nil {
		return "", errors.Wrap(err, "failed to format value")
	}
	return key + string(KVSep) + val, nil
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

func FormatKey(key keyval.Key) (string, error) {
	dir, err := FormatDirectory(key.Directory)
	if err != nil {
		return "", errors.Wrap(err, "failed to format directory")
	}
	tup, err := FormatTuple(key.Tuple)
	if err != nil {
		return "", errors.Wrap(err, "failed to format tuple")
	}
	return dir + tup, nil
}

func ParseDirectory(str string) (keyval.Directory, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != DirSep {
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

func FormatDirectory(dir keyval.Directory) (string, error) {
	var out strings.Builder
	for i, d := range dir {
		out.WriteRune(DirSep)
		switch d := d.(type) {
		case string:
			out.WriteString(d)
		case keyval.Variable:
			out.WriteString(FormatVariable(d))
		default:
			return "", errors.Errorf("failed to format %s element - '%v' (%T)", ordinal(i), d, d)
		}
	}
	return out.String(), nil
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

	var tup keyval.Tuple
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

func FormatTuple(tup keyval.Tuple) (string, error) {
	if len(tup) == 0 {
		return "", nil
	}

	var out strings.Builder
	out.WriteRune(TupStart)
	for i, t := range tup {
		if i != 0 {
			out.WriteRune(TupSep)
		}
		switch t := t.(type) {
		case keyval.Tuple:
			str, err := FormatTuple(t)
			if err != nil {
				return "", errors.Wrapf(err, "failed to format tuple at %s element", ordinal(i))
			}
			out.WriteString(str)
		default:
			str, err := FormatData(t)
			if err != nil {
				return "", errors.Wrapf(err, "failed to format data at %s element", ordinal(i))
			}
			out.WriteString(str)
		}
	}
	out.WriteRune(TupEnd)
	return out.String(), nil
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

func FormatData(in interface{}) (string, error) {
	switch in := in.(type) {
	case nil:
		return Nil, nil
	case bool:
		if in {
			return True, nil
		} else {
			return False, nil
		}
	case keyval.Variable:
		return FormatVariable(in), nil
	case string:
		return FormatString(in), nil
	case tuple.UUID:
		return FormatUUID(in), nil
	default:
		str, err := FormatNumber(in)
		return str, errors.Wrap(err, "failed to format as number")
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

func FormatVariable(in keyval.Variable) string {
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
	if str[0] != StrStart {
		return "", errors.New("strings must start with double quotes")
	}
	if str[len(str)-1] != StrEnd {
		return "", errors.New("strings must end with double quotes")
	}
	return str[1 : len(str)-1], nil
}

func FormatString(in string) string {
	var out strings.Builder
	out.WriteRune(StrStart)
	out.WriteString(in)
	out.WriteRune(StrEnd)
	return out.String()
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

func FormatUUID(in tuple.UUID) string {
	var out strings.Builder
	out.WriteString(hex.EncodeToString(in[:4]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[4:6]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[6:8]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[8:10]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[10:]))
	return out.String()
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

func FormatNumber(in interface{}) (string, error) {
	switch in := in.(type) {
	// Int
	case int64:
		return strconv.FormatInt(in, 10), nil
	case int:
		return strconv.FormatInt(int64(in), 10), nil

	// Uint
	case uint64:
		return strconv.FormatUint(in, 10), nil
	case uint:
		return strconv.FormatUint(uint64(in), 10), nil

	// Float
	case float64:
		return strconv.FormatFloat(in, 'g', 10, 64), nil
	case float32:
		return strconv.FormatFloat(float64(in), 'g', 10, 64), nil

	default:
		return "", errors.Errorf("unexpected input %v (%T)", in, in)
	}
}

func ParseValue(in string) (keyval.Value, error) {
	if len(in) == 0 {
		return nil, errors.New("input is empty")
	}
	if in == Clear {
		return keyval.Clear{}, nil
	}
	if in[0] == '(' {
		out, err := ParseTuple(in)
		return out, errors.Wrap(err, "failed to parse as tuple")
	}
	out, err := ParseData(in)
	return out, errors.Wrap(err, "failed to parse as data")
}

func FormatValue(in keyval.Value) (string, error) {
	switch in := in.(type) {
	case keyval.Clear:
		return Clear, nil
	case keyval.Tuple:
		str, err := FormatTuple(in)
		return str, errors.Wrap(err, "failed to format as tuple")
	default:
		str, err := FormatData(in)
		return str, errors.Wrap(err, "failed to format as data")
	}
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
