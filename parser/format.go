package parser

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

func FormatKeyValue(kv q.KeyValue) (string, error) {
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

func FormatKey(key q.Key) (string, error) {
	dir, err := FormatDirectory(key.Directory)
	if err != nil {
		return "", errors.Wrap(err, "failed to format directory")
	}
	if len(key.Tuple) == 0 {
		return dir, nil
	}
	tup, err := FormatTuple(key.Tuple)
	if err != nil {
		return "", errors.Wrap(err, "failed to format tuple")
	}
	return dir + tup, nil
}

func FormatDirectory(dir q.Directory) (string, error) {
	var out strings.Builder
	for i, d := range dir {
		out.WriteRune(DirSep)
		switch d := d.(type) {
		case string:
			out.WriteString(d)
		case q.Variable:
			out.WriteString(FormatVariable(d))
		default:
			return "", errors.Errorf("failed to format %s element - '%v' (%T)", ordinal(i), d, d)
		}
	}
	return out.String(), nil
}

func FormatTuple(tup q.Tuple) (string, error) {
	var out strings.Builder
	out.WriteRune(TupStart)
	for i, t := range tup {
		if i != 0 {
			out.WriteRune(TupSep)
		}
		switch t := t.(type) {
		case q.Tuple:
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
	case q.Variable:
		return FormatVariable(in), nil
	case string:
		return FormatString(in)
	case []byte:
		return FormatString(in)
	case tuple.UUID:
		return FormatUUID(in), nil
	default:
		str, err := FormatNumber(in)
		return str, errors.Wrap(err, "failed to format as number")
	}
}

func FormatVariable(in q.Variable) string {
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

func FormatString(in interface{}) (string, error) {
	var out strings.Builder
	out.WriteRune(StrStart)
	switch in := in.(type) {
	case string:
		out.WriteString(in)
	case []byte:
		out.WriteString(StrHex)
		out.WriteString(hex.EncodeToString(in))
	default:
		return "", errors.Errorf("failed to format '%v' (%T)", in, in)
	}
	out.WriteRune(StrEnd)
	return out.String(), nil
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

func FormatValue(in q.Value) (string, error) {
	switch in := in.(type) {
	case q.Clear:
		return Clear, nil
	case q.Tuple:
		str, err := FormatTuple(in)
		return str, errors.Wrap(err, "failed to format as tuple")
	default:
		str, err := FormatData(in)
		return str, errors.Wrap(err, "failed to format as data")
	}
}
