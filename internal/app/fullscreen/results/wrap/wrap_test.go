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
			4,
			"foobar",
			[]string{"foob", "ar"},
		},
		{
			"break many words",
			4,
			"funder i knew",
			[]string{"fund", "er i", "knew"},
		},
		{
			"wrap between two words",
			3,
			"foo bar",
			[]string{"foo", "bar"},
		},
		{
			"wrap between many words",
			4,
			"foo bar bing baz",
			[]string{"foo ", "bar ", "bing", "baz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.output, Wrap(tt.input, tt.limit))
		})
	}
}
