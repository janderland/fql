package results

import (
	"fmt"
	"strings"
	"testing"

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
	x := New()
	x.Height(1)

	x.Push(errors.New("error"))
	require.Equal(t, "1  ERR! error\n", x.View())

	x.Reset()
	x.Push("string")
	require.Equal(t, "1  # string\n", x.View())

	x.Reset()
	x.Push(keyval.KeyValue{
		Key: keyval.Key{
			Directory: keyval.Directory{keyval.String("dir")},
			Tuple:     keyval.Tuple{keyval.Int(23)},
		},
		Value: keyval.Int(10),
	})
	require.Equal(t, "1  /dir{23}=10\n", x.View())

	// TODO: Actually mock the GetPath method.
	x.Reset()
	x.Push(facade.NewNilDirectorySubspace())
	require.Equal(t, "1  \n", x.View())

	x.Reset()
	x.Push(stream.KeyValErr{
		KV: keyval.KeyValue{
			Key: keyval.Key{
				Directory: keyval.Directory{keyval.String("dir")},
				Tuple:     keyval.Tuple{keyval.Int(23)},
			},
			Value: keyval.Int(10),
		},
	})
	require.Equal(t, "1  /dir{23}=10\n", x.View())

	x.Reset()
	x.Push(stream.KeyValErr{
		Err: errors.New("error"),
	})
	require.Equal(t, "1  ERR! error\n", x.View())

	// TODO: Actually mock the GetPath method.
	x.Reset()
	x.Push(stream.DirErr{
		Dir: facade.NewNilDirectorySubspace(),
	})
	require.Equal(t, "1  \n", x.View())

	x.Reset()
	x.Push(stream.DirErr{
		Err: errors.New("error"),
	})
	require.Equal(t, "1  ERR! error\n", x.View())

	x.Reset()
	x.Push([]uint8{})
	require.Equal(t, "1  ERR! unexpected []uint8\n", x.View())
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
