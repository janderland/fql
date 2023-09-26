package wrap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		input  string
		output []string
	}{
		{
			"break single word",
			3,
			"foobar",
			[]string{"foo", "bar"},
		},
		{
			"wrap between words",
			3,
			"foo bar",
			[]string{"foo", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.output, Wrap(tt.input, tt.limit))
		})
	}
}
