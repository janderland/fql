package parser

import (
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

var (
	tkKVSep    = tokenWithKind{TokenKindKVSep, string(KVSep)}
	tkDirSep   = tokenWithKind{TokenKindDirSep, string(DirSep)}
	tkTupStart = tokenWithKind{TokenKindTupStart, string(TupStart)}
	tkTupEnd   = tokenWithKind{TokenKindTupEnd, string(TupEnd)}
	tkTupSep   = tokenWithKind{TokenKindTupSep, string(TupSep)}
	tkStrMark  = tokenWithKind{TokenKindStrMark, string(StrMark)}
)

type tokenWithKind struct {
	kind  TokenKind
	token string
}

func TestScanner(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		tokens []tokenWithKind
	}{
		{
			name:  "dirs",
			input: "/my\r\n/dir\t ",
			tokens: []tokenWithKind{
				tkDirSep,
				{TokenKindOther, "my\r\n"},
				tkDirSep,
				{TokenKindOther, "dir\t "},
			},
		},
		{
			name:  "tuples",
			input: "{\"something\"\r, \t22.88e0,- 88  \n}",
			tokens: []tokenWithKind{
				tkTupStart,
				tkStrMark,
				{TokenKindOther, "something"},
				tkStrMark,
				{TokenKindNewLine, "\r"},
				tkTupSep,
				{TokenKindWhitespace, " \t"},
				{TokenKindOther, "22.88e0"},
				tkTupSep,
				{TokenKindOther, "-"},
				{TokenKindWhitespace, " "},
				{TokenKindOther, "88"},
				{TokenKindNewLine, "  \n"},
				tkTupEnd,
			},
		},
		{
			name:  "key-value",
			input: "/my \t/dir\r\n{ \"hi world\" ,\n 88-212 = {, \t",
			tokens: []tokenWithKind{
				tkDirSep,
				{TokenKindOther, "my \t"},
				tkDirSep,
				{TokenKindOther, "dir\r\n"},
				tkTupStart,
				{TokenKindWhitespace, " "},
				tkStrMark,
				{TokenKindOther, "hi world"},
				tkStrMark,
				{TokenKindWhitespace, " "},
				tkTupSep,
				{TokenKindNewLine, "\n "},
				{TokenKindOther, "88-212"},
				{TokenKindWhitespace, " "},
				tkKVSep,
				{TokenKindWhitespace, " "},
				tkTupStart,
				tkTupSep,
				{TokenKindWhitespace, " \t"},
			},
		},
		{
			name:  "escape",
			input: "/how \\a\n /wow { \"tens \\\\ \"",
			tokens: []tokenWithKind{
				tkDirSep,
				{TokenKindOther, "how "},
				{TokenKindEscape, "\\a"},
				{TokenKindOther, "\n "},
				tkDirSep,
				{TokenKindOther, "wow "},
				tkTupStart,
				{TokenKindWhitespace, " "},
				tkStrMark,
				{TokenKindOther, "tens "},
				{TokenKindEscape, "\\\\"},
				{TokenKindOther, " "},
				tkStrMark,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(test.input))
			var tokens []tokenWithKind

			for {
				kind, err := s.Scan()
				require.NoError(t, err)
				if kind == TokenKindEnd {
					break
				}

				tokens = append(tokens, tokenWithKind{
					kind:  kind,
					token: s.Token(),
				})
			}

			require.Equal(t, test.tokens, tokens)
		})
	}
}

func TestErrRecovery(t *testing.T) {
	s := NewScanner(&badReader{})
	_, err := s.Scan()
	require.Error(t, err)
}

type badReader struct{}

func (x *badReader) Read(_ []byte) (int, error) {
	return 0, errors.New("this reader always fails")
}
