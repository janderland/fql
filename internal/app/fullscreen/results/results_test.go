package results

import (
	"fmt"
	"strings"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/engine/stream"
	"github.com/janderland/fdbq/keyval"
)

func TestHeight(t *testing.T) {
	x := setup()

	require.Empty(t, x.View())

	x.Height(50)
	require.Equal(t, 50, strings.Count(x.View(), "\n"))
}

func TestSingleLine(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{
			errors.New("error"),
			"1  ERR! error\n",
		},
		{
			"string",
			"1  # string\n",
		},
		{
			keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.Directory{keyval.String("dir")},
					Tuple:     keyval.Tuple{keyval.Int(23)},
				},
				Value: keyval.Int(10),
			},
			"1  /dir{23}=10\n",
		},
		{
			dir([]string{"dir"}),
			"1  /dir\n",
		},
		{
			stream.KeyValErr{
				KV: keyval.KeyValue{
					Key: keyval.Key{
						Directory: keyval.Directory{keyval.String("dir")},
						Tuple:     keyval.Tuple{keyval.Int(23)},
					},
					Value: keyval.Int(10),
				},
			},
			"1  /dir{23}=10\n",
		},
		{
			stream.KeyValErr{
				Err: errors.New("error"),
			},
			"1  ERR! error\n",
		},
		{
			stream.DirErr{
				Dir: dir([]string{"dir"}),
			},
			"1  /dir\n",
		},
		{
			stream.DirErr{
				Err: errors.New("error"),
			},
			"1  ERR! error\n",
		},
		{
			[]uint8{},
			"1  ERR! unexpected []uint8\n",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.input), func(t *testing.T) {
			x := New()
			x.Height(1)

			x.Push(test.input)
			require.Equal(t, test.expected, x.View())
		})
	}
}

func TestScroll(t *testing.T) {
	t.Run("up", func(t *testing.T) {
		x := setup()
		x.Height(5)

		var expected strings.Builder
		for i := 96; i <= 100; i++ {
			expected.WriteString(fmt.Sprintf("%d  # %d\n", i, i))
		}
		require.Equal(t, expected.String(), x.View())
	})
}

func setup() Model {
	x := New()
	for i := 1; i <= 100; i++ {
		x.Push(fmt.Sprintf("%d", i))
	}
	return x
}

func dir(path []string) directory.DirectorySubspace {
	return &mockDir{facade.NewNilDirectorySubspace(), path}
}

type mockDir struct {
	directory.DirectorySubspace
	path []string
}

func (x *mockDir) GetPath() []string {
	return x.path
}
