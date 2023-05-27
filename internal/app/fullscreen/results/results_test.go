package results

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeight(t *testing.T) {
	x := New()
	for i := 0; i < 100; i++ {
		x.Push(fmt.Sprintf("%d", i))
	}

	require.Empty(t, x.View())

	x.Height(50)
	require.Equal(t, 50, countLines(x.View()))
}

func countLines(str string) int {
	return strings.Count(str, "\n")
}
