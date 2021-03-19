package keyval

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

func ValToString(val Value) (string, error) {
	b, ok := val.([]byte)
	if !ok {
		return "", errors.New("is not []byte")
	}
	return string(b), nil
}

func ValFromString(str string) Value {
	return []byte(str)
}

func ValToInt(val Value) (int64, error) {
	i, err := ValToUint(val)
	return int64(i), err
}

func ValFromInt(i int64) Value {
	return ValFromUint(uint64(i))
}

func ValToUint(val Value) (uint64, error) {
	b, ok := val.([]byte)
	if !ok {
		return 0, errors.New("not []byte")
	}
	if len(b) > 8 {
		return 0, errors.New("larger than 8 bytes")
	}

	// Ensure b is 8 bytes.
	b = append(b, make([]byte, 8-len(b))...)

	return binary.LittleEndian.Uint64(b), nil
}

func ValFromUint(i uint64) Value {
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, i)
	return val
}
