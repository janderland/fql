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
			"break word",
			4,
			"foobar",
			[]string{"foob", "ar"},
		},
		{
			"wrap between two words",
			3,
			"foo bar",
			[]string{"foo", "bar"},
		},
		{
			"wrap between many words",
			7,
			"foo bar bing baz",
			[]string{"foo bar", "bing ", "baz"},
		},
		{
			"break and wrap words",
			4,
			"funder i knew go there",
			[]string{"fund", "er i", "knew", "go ", "ther", "e"},
		},
		{
			"ignore ascii escape codes",
			4,
			"ba\x1b[1;31mll",
			[]string{"ba\x1b[1;31mll"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.output, Wrap(tt.input, tt.limit))
		})
	}
}
