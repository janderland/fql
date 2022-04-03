package parser

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	tokenKVSep    = token{TokenKindKVSep, string(KVSep)}
	tokenDirSep   = token{TokenKindDirSep, string(DirSep)}
	tokenTupStart = token{TokenKindTupStart, string(TupStart)}
	tokenTupEnd   = token{TokenKindTupEnd, string(TupEnd)}
	tokenTupSep   = token{TokenKindTupSep, string(TupSep)}
	tokenStrMark  = token{TokenKindStrMark, string(StrMark)}
)

type token struct {
	kind  TokenKind
	token string
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
				tokenDirSep,
				{TokenKindOther, "my\r\n"},
				tokenDirSep,
				{TokenKindOther, "dir\t "},
			},
		},
		{
			name:  "tuples",
			input: "{\"something\"\r, \t22.88e0,- 88  \n}",
			tokens: []token{
				tokenTupStart,
				tokenStrMark,
				{TokenKindOther, "something"},
				tokenStrMark,
				{TokenKindNewline, "\r"},
				tokenTupSep,
				{TokenKindWhitespace, " \t"},
				{TokenKindOther, "22.88e0"},
				tokenTupSep,
				{TokenKindOther, "-"},
				{TokenKindWhitespace, " "},
				{TokenKindOther, "88"},
				{TokenKindNewline, "  \n"},
				tokenTupEnd,
			},
		},
		{
			name:  "key-value",
			input: "/my \t/dir\r\n{ \"hi world\" ,\n 88-212 = {, \t",
			tokens: []token{
				tokenDirSep,
				{TokenKindOther, "my \t"},
				tokenDirSep,
				{TokenKindOther, "dir\r\n"},
				tokenTupStart,
				{TokenKindWhitespace, " "},
				tokenStrMark,
				{TokenKindOther, "hi world"},
				tokenStrMark,
				{TokenKindWhitespace, " "},
				tokenTupSep,
				{TokenKindNewline, "\n "},
				{TokenKindOther, "88-212"},
				{TokenKindWhitespace, " "},
				tokenKVSep,
				{TokenKindWhitespace, " "},
				tokenTupStart,
				tokenTupSep,
				{TokenKindWhitespace, " \t"},
			},
		},
		{
			name:  "escape",
			input: "/how \\a\n /wow { \"tens \\\\ \"",
			tokens: []token{
				tokenDirSep,
				{TokenKindOther, "how "},
				{TokenKindEscape, "\\a"},
				{TokenKindOther, "\n "},
				tokenDirSep,
				{TokenKindOther, "wow "},
				tokenTupStart,
				{TokenKindWhitespace, " "},
				tokenStrMark,
				{TokenKindOther, "tens "},
				{TokenKindEscape, "\\\\"},
				{TokenKindOther, " "},
				tokenStrMark,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(test.input))
			var tokens []token

			for {
				kind, err := s.Scan()
				require.NoError(t, err)
				if kind == TokenKindEnd {
					break
				}

				tokens = append(tokens, token{
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
