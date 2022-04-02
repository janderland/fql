package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const (
	// runesWhitespace contains the characters allowed to be
	// in a TokenKindWhitespace token.
	runesWhitespace = "\t "

	// runesNewline, together with runesWhitespace, contains the
	// characters allowed to be in a TokenKindNewLine token.
	runesNewline = "\n\r"
)

// TokenKind represents the kind of token read during a call to Scanner.Scan.
type TokenKind int

const (
	// TokenKindUnassigned is used to identify a TokenKind variable
	// which hasn't been assigned to yet. For this purpose, it must
	// be the zero-value of its type.
	TokenKindUnassigned TokenKind = iota

	// TokenKindWhitespace identifies a token which only contains
	// characters found in the runesWhitespace constant.
	TokenKindWhitespace

	// TokenKindNewLine identifies a token which only contains
	// characters found in the runesWhitespace or runesNewline
	// constants.
	TokenKindNewLine

	// TokenKindEscape identifies a 2-character token which always
	// starts with the Escape character.
	TokenKindEscape

	// TokenKindOther identifies all other possible tokens which are
	// not identified by the given TokenKind constants. This kind of
	// token is used to represent directory names, value types, and
	// data elements (numbers, strings, UUIDs, etc...).
	TokenKindOther

	// TokenKindEnd is returned from Scanner.Scan when the wrapped
	// io.Reader has been read to completion.
	TokenKindEnd

	// TokenKindKVSep identifies a token equal to KVSep.
	TokenKindKVSep

	// TokenKindDirSep identifies a token equal to DirSep.
	TokenKindDirSep

	// TokenKindTupStart identifies a token equal to TupStart.
	TokenKindTupStart

	// TokenKindTupEnd identifies a token equal to TupEnd.
	TokenKindTupEnd

	// TokenKindTupSep identifies a token equal to TupSep.
	TokenKindTupSep

	// TokenKindVarStart identifies a token equal to VarStart.
	TokenKindVarStart

	// TokenKindVarEnd identifies a token equal to VarEnd.
	TokenKindVarEnd

	// TokenKindVarSep identifies a token equal to VarSep.
	TokenKindVarSep

	// TokenKindStrMark identifies a token equal to StrMark.
	TokenKindStrMark
)

var specialKindByRune = map[rune]TokenKind{
	KVSep:    TokenKindKVSep,
	DirSep:   TokenKindDirSep,
	TupStart: TokenKindTupStart,
	TupEnd:   TokenKindTupEnd,
	TupSep:   TokenKindTupSep,
	VarStart: TokenKindVarStart,
	VarEnd:   TokenKindVarEnd,
	VarSep:   TokenKindVarSep,
	StrMark:  TokenKindStrMark,
}

type scannerState int

const (
	scannerStateUnassigned scannerState = iota
	scannerStateWhitespace
	scannerStateNewline
	scannerStateDirPart
	scannerStateString
	scannerStateOther
)

var primaryKindByState = map[scannerState]TokenKind{
	scannerStateWhitespace: TokenKindWhitespace,
	scannerStateNewline:    TokenKindNewLine,
	scannerStateDirPart:    TokenKindOther,
	scannerStateString:     TokenKindOther,
	scannerStateOther:      TokenKindOther,
}

// Scanner splits the bytes from an io.Reader into tokens
// for a Parser. For each call to the Scan method, a token
// worth of bytes is read from the io.Reader and the kind
// of token is returned. After a call to Scan, the Token
// method may be called to obtain the token string.
type Scanner struct {
	source *bufio.Reader
	token  *strings.Builder
	state  scannerState
	escape bool
}

// NewScanner creates a Scanner which reads from the given io.Reader.
func NewScanner(src io.Reader) Scanner {
	var token strings.Builder
	return Scanner{
		source: bufio.NewReader(src),
		token:  &token,
		state:  scannerStateWhitespace,
	}
}

func (x *Scanner) Token() string {
	return x.token.String()
}

func (x *Scanner) Scan() (kind TokenKind, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				kind = TokenKindUnassigned
				err = e
				return
			}
			panic(r)
		}
	}()

	x.token.Reset()

	for {
		r, eof := x.read()
		if eof {
			if x.token.Len() > 0 {
				return primaryKindByState[x.state], nil
			}
			return TokenKindEnd, nil
		}

		if x.escape {
			x.escape = false
			x.append(r)
			return TokenKindEscape, nil
		} else if r == Escape {
			if x.token.Len() > 0 {
				x.unread()
				return primaryKindByState[x.state], nil
			}
			x.escape = true
			x.append(r)
			continue
		}

		if kind, ok := specialKindByRune[r]; ok {
			newState := scannerStateUnassigned

			switch r {
			case DirSep:
				switch x.state {
				case scannerStateString:
					break
				default:
					newState = scannerStateDirPart
				}

			case StrMark:
				switch x.state {
				case scannerStateDirPart:
					break
				case scannerStateString:
					newState = scannerStateWhitespace
				default:
					newState = scannerStateString
				}

			default:
				switch x.state {
				case scannerStateString:
					break
				default:
					newState = scannerStateWhitespace
				}
			}

			if newState != scannerStateUnassigned {
				if x.token.Len() > 0 {
					x.unread()
					return primaryKindByState[x.state], nil
				}
				x.state = newState
			}

			x.append(r)
			return kind, nil
		}

		if strings.ContainsRune(runesWhitespace, r) {
			switch x.state {
			case scannerStateOther:
				x.unread()
				kind := primaryKindByState[x.state]
				x.state = scannerStateWhitespace
				return kind, nil

			default:
				x.append(r)
				continue
			}
		}

		if strings.ContainsRune(runesNewline, r) {
			switch x.state {
			case scannerStateWhitespace:
				x.state = scannerStateNewline
				x.append(r)
				continue

			case scannerStateOther:
				x.unread()
				kind := primaryKindByState[x.state]
				x.state = scannerStateNewline
				return kind, nil

			default:
				x.append(r)
				continue
			}
		}

		switch x.state {
		case scannerStateWhitespace, scannerStateNewline:
			if x.token.Len() == 0 {
				x.state = scannerStateOther
				x.append(r)
				continue
			}
			x.unread()
			kind := primaryKindByState[x.state]
			x.state = scannerStateOther
			return kind, nil

		default:
			x.append(r)
			continue
		}
	}
}

func (x *Scanner) append(r rune) {
	_, err := x.token.WriteRune(r)
	if err != nil {
		panic(errors.Wrap(err, "failed to append rune"))
	}
}

func (x *Scanner) read() (rune, bool) {
	r, _, err := x.source.ReadRune()
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		panic(errors.Wrap(err, "failed to read rune"))
	}
	return r, false
}

func (x *Scanner) unread() {
	err := x.source.UnreadRune()
	if err != nil {
		panic(errors.Wrap(err, "failed to unread rune"))
	}
}
