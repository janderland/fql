package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const (
	runesWhitespace = "\t "
	runesNewline    = "\n\r"
)

type TokenKind int

const (
	TokenKindInvalid TokenKind = iota
	TokenKindKVSep
	TokenKindDirSep
	TokenKindTupStart
	TokenKindTupEnd
	TokenKindTupSep
	TokenKindVarStart
	TokenKindVarEnd
	TokenKindVarSep
	TokenKindStrMark
	TokenKindWhitespace
	TokenKindNewLine
	TokenKindOther
	TokenKindEnd
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
	ScannerStateInvalid scannerState = iota
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

type Scanner struct {
	reader *bufio.Reader
	token  strings.Builder
	state  scannerState
}

func NewScanner(rd io.Reader) Scanner {
	return Scanner{
		reader: bufio.NewReader(rd),
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
				kind = TokenKindInvalid
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
			if x.token.Len() == 0 {
				return TokenKindEnd, nil
			}
			return primaryKindByState[x.state], nil
		}

		if kind, ok := specialKindByRune[r]; ok {
			newState := ScannerStateInvalid

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

			if newState != ScannerStateInvalid {
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
	r, _, err := x.reader.ReadRune()
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		panic(errors.Wrap(err, "failed to read rune"))
	}
	return r, false
}

func (x *Scanner) unread() {
	err := x.reader.UnreadRune()
	if err != nil {
		panic(errors.Wrap(err, "failed to unread rune"))
	}
}
