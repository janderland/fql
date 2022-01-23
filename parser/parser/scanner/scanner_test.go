package scanner

import (
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/janderland/fdbq/parser/parser"
	"github.com/stretchr/testify/require"
)

const (
	KVSep    = string(parser.KVSep)
	DirSep   = string(parser.DirSep)
	TupStart = string(parser.TupStart)
	TupEnd   = string(parser.TupEnd)
	TupSep   = string(parser.TupSep)
	StrMark  = string(parser.StrMark)
)

type token struct {
	kind TokenKind
	str  string
}

func TestScanner(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		tokens []token
	}{
		{
			name:  "dirs",
			input: "/my\r\n/dir\t ",
			tokens: []token{
				{TokenDirSep, DirSep},
				{TokenOther, "my\r\n"},
				{TokenDirSep, DirSep},
				{TokenOther, "dir\t "},
			},
		},
		{
			name:  "tuples",
			input: "{\"something\"\r, \t22.88e0,- 88  \n}",
			tokens: []token{
				{TokenTupStart, TupStart},
				{TokenStrMark, StrMark},
				{TokenOther, "something"},
				{TokenStrMark, StrMark},
				{TokenNewLine, "\r"},
				{TokenTupSep, TupSep},
				{TokenWhitespace, " \t"},
				{TokenOther, "22.88e0"},
				{TokenTupSep, TupSep},
				{TokenOther, "-"},
				{TokenWhitespace, " "},
				{TokenOther, "88"},
				{TokenNewLine, "  \n"},
				{TokenTupEnd, TupEnd},
			},
		},
		{
			name:  "key-value",
			input: "/my \t/dir\r\n{ \"hi world\" ,\n 88-212 = {, \t",
			tokens: []token{
				{TokenDirSep, DirSep},
				{TokenOther, "my \t"},
				{TokenDirSep, DirSep},
				{TokenOther, "dir\r\n"},
				{TokenTupStart, TupStart},
				{TokenWhitespace, " "},
				{TokenStrMark, StrMark},
				{TokenOther, "hi world"},
				{TokenStrMark, StrMark},
				{TokenWhitespace, " "},
				{TokenTupSep, TupSep},
				{TokenNewLine, "\n "},
				{TokenOther, "88-212"},
				{TokenWhitespace, " "},
				{TokenKVSep, KVSep},
				{TokenWhitespace, " "},
				{TokenTupStart, TupStart},
				{TokenTupSep, TupSep},
				{TokenWhitespace, " \t"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := New(strings.NewReader(test.input))
			var tokens []token

			for {
				kind, err := s.Scan()
				require.NoError(t, err)
				if kind == TokenEnd {
					break
				}

				tokens = append(tokens, token{
					kind: kind,
					str:  s.Token(),
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
