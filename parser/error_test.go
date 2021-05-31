package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

func TestParseError_Error(t *testing.T) {
	wrap1 := newErrWrapper("one", "1234")
	wrap2 := newErrWrapper("two", "2345")
	wrap3 := newErrWrapper("three", "6789")
	err := wrap1(wrap2(wrap3(errors.New("root error"))))
	assert.Error(t, err)
	t.Log(err.Error())
}
