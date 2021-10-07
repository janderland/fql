package parser

import (
	"encoding/hex"
	"strconv"
	"strings"

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
		return nil, errors.Errorf("missing '%c' separator between key and value", KVSep)
	} else if len(parts) > 2 {
		return nil, errors.Errorf("contains multiple '%c'", KVSep)
	}

	keyStr := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	key, err := ParseKey(keyStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse key - '%s'", keyStr)
	}
	value, err := ParseValue(valueStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse value - '%s'", valueStr)
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
	if i := strings.Index(str, string(TupStart)); i == -1 {
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
			return nil, errors.Wrapf(err, "failed to parse directory - '%s'", directoryStr)
		}
	}
	if len(tupleStr) > 0 {
		key.Tuple, err = ParseTuple(tupleStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse tuple - '%s'", tupleStr)
		}
	}
	return key, nil
}

func ParseDirectory(str string) (q.Directory, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] != DirSep {
		return nil, errors.Errorf("directory path must start with a '%c'", DirSep)
	}
	if str[len(str)-1] == DirSep {
		return nil, errors.Errorf("directory path shouldn't have a trailing '%c'", DirSep)
	}

	var directory q.Directory
	for i, part := range strings.Split(str[1:], string(DirSep)) {
		part = strings.TrimSpace(part)
		element, err := ParsePathElement(part)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s element - '%s", ordinal(i+1), part)
		}
		directory = append(directory, element)
	}
	return directory, nil
}

func ParsePathElement(str string) (q.DirElement, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str[0] == VarStart {
		variable, err := ParseVariable(str)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse as variable")
		}
		return variable, nil
	} else {
		return q.String(str), nil
	}
}

func ParseTuple(str string) (q.Tuple, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}

	if str[0] != TupStart {
		return nil, errors.Errorf("must start with a '%c'", TupStart)
	}
	if str[len(str)-1] != TupEnd {
		return nil, errors.Errorf("must end with a '%c'", TupEnd)
	}

	tup := q.Tuple{}

	str = str[1 : len(str)-1]
	if len(str) == 0 {
		return tup, nil
	}

	parts := strings.Split(str, string(TupSep))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		element, err := ParseTupleElement(part)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s element - '%s'", ordinal(i+1), part)
		}
		if _, isMaybeMore := element.(q.MaybeMore); isMaybeMore && i != len(parts)-1 {
			return nil, errors.New("'...' should only appear as the last element of a tuple")
		}
		tup = append(tup, element)
	}
	return tup, nil
}

func ParseTupleElement(str string) (q.TupElement, error) {
	if len(str) == 0 {
		return nil, errors.New("element is empty")
	}

	if str[0] == TupStart {
		tup, err := ParseTuple(str)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse as tuple")
		}
		return tup, nil
	}
	if str == MaybeMore {
		return q.MaybeMore{}, nil
	}
	data, err := ParseData(str)
	if err != nil {
		return nil, err
	}
	return data.(q.TupElement), nil
}

func ParseData(str string) (interface{}, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == Nil {
		return q.Nil{}, nil
	}
	if str == True {
		return q.Bool(true), nil
	}
	if str == False {
		return q.Bool(false), nil
	}
	if str[0] == VarStart {
		data, err := ParseVariable(str)
		return data, errors.Wrap(err, "failed to parse as variable")
	}
	if str[0] == StrStart {
		data, err := ParseString(str)
		return data, errors.Wrap(err, "failed to parse as string")
	}
	if strings.HasPrefix(str, HexStart) {
		data, err := ParseHex(str)
		return data, errors.Wrap(err, "failed to parse as hex string")
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
	if str[0] != VarStart {
		return nil, errors.Errorf("must start with '%c'", VarStart)
	}
	if str[len(str)-1] != VarEnd {
		return nil, errors.Errorf("must end with '%c'", VarEnd)
	}

	var variable q.Variable
	if typeUnionStr := str[1 : len(str)-1]; len(typeUnionStr) > 0 {
	loop:
		for i, typeStr := range strings.Split(typeUnionStr, string(VarSep)) {
			typeStr = strings.TrimSpace(typeStr)
			for _, v := range q.AllTypes() {
				if string(v) == typeStr {
					variable = append(variable, q.ValueType(typeStr))
					continue loop
				}
			}
			return nil, errors.Errorf("failed to parse %s type - '%s': invalid type", ordinal(i+1), typeStr)
		}
	}
	return variable, nil
}

func ParseString(str string) (q.String, error) {
	if len(str) == 0 {
		return "", errors.New("input is empty")
	}
	if str[0] != StrStart {
		return "", errors.New("must start with double quotes")
	}
	if str[len(str)-1] != StrEnd {
		return "", errors.New("must end with double quotes")
	}
	return q.String(str[1 : len(str)-1]), nil
}

func ParseHex(str string) (q.Bytes, error) {
	if !strings.HasPrefix(str, HexStart) {
		return nil, errors.Errorf("expected '%s' prefix", HexStart)
	}
	str = str[len(HexStart):]
	if len(str)%2 != 0 {
		return nil, errors.New("expected even number of hex digits")
	}
	return hex.DecodeString(str)
}

func ParseUUID(str string) (q.UUID, error) {
	if len(str) == 0 {
		return q.UUID{}, errors.New("input is empty")
	}

	groups := strings.Split(str, "-")
	checkLen := func(i int, expLen int) error {
		if len(groups[i]) != expLen {
			return errors.Errorf("the %s group should contain %d characters rather than %d", ordinal(i+1), expLen, len(groups[i]))
		}
		return nil
	}
	if err := checkLen(0, 8); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(1, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(2, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(3, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(4, 12); err != nil {
		return q.UUID{}, err
	}

	var uuid q.UUID
	_, err := hex.Decode(uuid[:], []byte(strings.ReplaceAll(str, "-", "")))
	if err != nil {
		return q.UUID{}, err
	}
	return uuid, nil
}

func ParseNumber(str string) (interface{}, error) {
	i, iErr := strconv.ParseInt(str, 10, 64)
	if iErr == nil {
		return q.Int(i), nil
	}
	u, uErr := strconv.ParseUint(str, 10, 64)
	if uErr == nil {
		return q.Uint(u), nil
	}
	f, fErr := strconv.ParseFloat(str, 64)
	if fErr == nil {
		return q.Float(f), nil
	}
	return nil, errors.New("invalid syntax")
}

func ParseValue(str string) (q.Value, error) {
	if len(str) == 0 {
		return nil, errors.New("input is empty")
	}
	if str == Clear {
		return q.Clear{}, nil
	}
	if str[0] == TupStart {
		out, err := ParseTuple(str)
		return out, errors.Wrap(err, "failed to parse as tuple")
	}
	val, err := ParseData(str)
	return val.(q.Value), err
}
