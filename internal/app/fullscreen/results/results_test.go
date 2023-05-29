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
	require.Equal(t, 50, lineCount(x.View()))
}

func TestReset(t *testing.T) {
	x := setup()
	x.Height(50)

	x.Reset()
	require.Empty(t, x.View())

	for i := 0; i < 10; i++ {
		x.Push("")
	}
	require.Equal(t, 10, lineCount(x.View()))
}

func TestSingleLine(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{
			errors.New("error"),
			"1  ERR! error",
		},
		{
			"string",
			"1  # string",
		},
		{
			keyval.KeyValue{
				Key: keyval.Key{
					Directory: keyval.Directory{keyval.String("dir")},
					Tuple:     keyval.Tuple{keyval.Int(23)},
				},
				Value: keyval.Int(10),
			},
			"1  /dir{23}=10",
		},
		{
			dir([]string{"dir"}),
			"1  /dir",
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
			"1  /dir{23}=10",
		},
		{
			stream.KeyValErr{
				Err: errors.New("error"),
			},
			"1  ERR! error",
		},
		{
			stream.DirErr{
				Dir: dir([]string{"dir"}),
			},
			"1  /dir",
		},
		{
			stream.DirErr{
				Err: errors.New("error"),
			},
			"1  ERR! error",
		},
		{
			[]uint8{},
			"1  ERR! unexpected []uint8",
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
	x := setup()
	x.Height(5)

	var expected strings.Builder
	for i := 96; i <= 100; i++ {
		if i != 96 {
			expected.WriteRune('\n')
		}
		expected.WriteString(fmt.Sprintf("%d  # %d", i, i))
	}
	require.Equal(t, expected.String(), x.View())

	x.scrollUp(10)
	expected.Reset()
	for i := 86; i <= 90; i++ {
		if i != 86 {
			expected.WriteRune('\n')
		}
		expected.WriteString(fmt.Sprintf("%d  # %d", i, i))
	}
	require.Equal(t, expected.String(), x.View())

	x.scrollDown(9)
	expected.Reset()
	for i := 95; i <= 99; i++ {
		if i != 95 {
			expected.WriteRune('\n')
		}
		expected.WriteString(fmt.Sprintf("%d  # %d", i, i))
	}
	require.Equal(t, expected.String(), x.View())
}

func setup() Model {
	x := New()
	for i := 1; i <= 100; i++ {
		x.Push(fmt.Sprintf("%d", i))
	}
	return x
}

func lineCount(str string) int {
	return len(strings.Split(str, "\n"))
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
