package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// TokenKind represents the kind of token read during a call to Scanner.Scan.
type TokenKind int

const (
	// TokenKindUnassigned is used to identify a TokenKind variable
	// which hasn't been assigned to yet. For this purpose, it must
	// be the zero-value of its type.
	TokenKindUnassigned TokenKind = iota

	// TokenKindWhitespace identifies a token which only contains
	// runes found in the runesWhitespace constant.
	TokenKindWhitespace

	// TokenKindNewline identifies a token which only contains
	// runes found in the runesWhitespace or runesNewline
	// constants.
	TokenKindNewline

	// TokenKindEscape identifies a 2-rune token which always
	// starts with the Escape rune.
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

type scannerState int

const (
	// scannerStateUnassigned is used to identify a scannerState variable
	// which hasn't been assigned to yet. For this purpose, it must
	// be the zero-value of its type.
	scannerStateUnassigned scannerState = iota

	// scannerStateWhitespace is the root state of the scanner. The scanner
	// remains in this state as long as the current token only contains runes
	// found in the runesWhitespace constant.
	scannerStateWhitespace

	// scannerStateNewline follows scannerStateWhitespace if any of the runes
	// found in the runesNewline constant are encountered. The scanner remains
	// in this state as long os the current token only contains runes found in
	// the runesWhitespace or runesNewline constants.
	scannerStateNewline

	// scannerStateDirPart follows any state, save for scannerStateString, if
	// a DirSep rune is encountered. During this state, the scanner doesn't
	// separate whitespace or newlines into separate tokens. The scanner remains
	// in this state until a TupStart is encountered.
	scannerStateDirPart

	// scannerStateString follows any state, save for scannerStateDirPart, if a
	// StrMark rune is encountered. During this state, the scanner doesn't
	// separate whitespace or newlines into separate tokens. The scanner remains
	// in this state until another StrMark is encountered.
	scannerStateString

	// scannerStateOther follows scannerStateWhitespace or scannerStateNewline
	// if a non-significant character is encountered. The scanner remains in
	// this state until a significant rune is encountered. Significant runes
	// include any of the single-rune constants from special.go, or any of
	// the runes in the constants runesWhitespace and runesNewline.
	scannerStateOther
)

const (
	// runesWhitespace contains the runes allowed to be
	// in a TokenKindWhitespace token.
	runesWhitespace = "\t "

	// runesNewline, together with runesWhitespace, contains the
	// runes allowed to be in a TokenKindNewline token.
	runesNewline = "\n\r"
)

// singleCharKindByRune contains all the TokenKind whose tokens must
// be a single rune, indexed by the rune which said tokens must contain.
// These kinds of tokens are all handled in a similar way and
// this map is be used by Scanner.Scan to identify them.
var singleCharKindByRune = map[rune]TokenKind{
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

// primaryKindByState maps a scannerState to the TokenKind usually returned
// by Scanner.Scan during a given state. If the scanner encounters an escape
// or any of the runes in singleCharKindByRune, then the Scanner.Scan method
// may return a different TokenKind than what this map provides.
var primaryKindByState = map[scannerState]TokenKind{
	scannerStateWhitespace: TokenKindWhitespace,
	scannerStateNewline:    TokenKindNewline,
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

// Token returns token obtained from the wrapped io.Reader by the last call to Scan.
func (x *Scanner) Token() string {
	return x.token.String()
}

// Scan reads a token from the wrapped io.Reader and returns the kind of token read.
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

		if kind, ok := singleCharKindByRune[r]; ok {
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
