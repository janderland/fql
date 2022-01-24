package parser

import (
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

var (
	tkKVSep    = tokenWithKind{TokenKVSep, string(KVSep)}
	tkDirSep   = tokenWithKind{TokenDirSep, string(DirSep)}
	tkTupStart = tokenWithKind{TokenTupStart, string(TupStart)}
	tkTupEnd   = tokenWithKind{TokenTupEnd, string(TupEnd)}
	tkTupSep   = tokenWithKind{TokenTupSep, string(TupSep)}
	tkStrMark  = tokenWithKind{TokenStrMark, string(StrMark)}
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
				{TokenOther, "my\r\n"},
				tkDirSep,
				{TokenOther, "dir\t "},
			},
		},
		{
			name:  "tuples",
			input: "{\"something\"\r, \t22.88e0,- 88  \n}",
			tokens: []tokenWithKind{
				tkTupStart,
				tkStrMark,
				{TokenOther, "something"},
				tkStrMark,
				{TokenNewLine, "\r"},
				tkTupSep,
				{TokenWhitespace, " \t"},
				{TokenOther, "22.88e0"},
				tkTupSep,
				{TokenOther, "-"},
				{TokenWhitespace, " "},
				{TokenOther, "88"},
				{TokenNewLine, "  \n"},
				tkTupEnd,
			},
		},
		{
			name:  "key-value",
			input: "/my \t/dir\r\n{ \"hi world\" ,\n 88-212 = {, \t",
			tokens: []tokenWithKind{
				tkDirSep,
				{TokenOther, "my \t"},
				tkDirSep,
				{TokenOther, "dir\r\n"},
				tkTupStart,
				{TokenWhitespace, " "},
				tkStrMark,
				{TokenOther, "hi world"},
				tkStrMark,
				{TokenWhitespace, " "},
				tkTupSep,
				{TokenNewLine, "\n "},
				{TokenOther, "88-212"},
				{TokenWhitespace, " "},
				tkKVSep,
				{TokenWhitespace, " "},
				tkTupStart,
				tkTupSep,
				{TokenWhitespace, " \t"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := New(strings.NewReader(test.input))
			var tokens []tokenWithKind

			for {
				kind, err := s.Scan()
				require.NoError(t, err)
				if kind == TokenEnd {
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
	s := New(&badReader{})
	_, err := s.Scan()
	require.Error(t, err)
}

type badReader struct{}

func (x *badReader) Read(_ []byte) (int, error) {
	return 0, errors.New("this reader always fails")
}
