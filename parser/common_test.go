package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrdinal(t *testing.T) {
	tests := []struct {
		name string
		in   int
		out  string
	}{
		{name: "zero", in: 0, out: "0th"},
		{name: "th", in: 8, out: "8th"},
		{name: "st", in: 21, out: "21st"},
		{name: "nd", in: 302, out: "302nd"},
		{name: "rd", in: 53, out: "53rd"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.out, ordinal(test.in))
		})
	}
}
