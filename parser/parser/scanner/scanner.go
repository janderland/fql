package scanner

import (
	"bufio"
	"io"
	"strings"

	"github.com/janderland/fdbq/parser/parser"
	"github.com/pkg/errors"
)

const (
	whitespace = "\t "
	newline    = "\t\n\r "
)

type TokenKind int

const (
	TokenInvalid TokenKind = iota

	TokenKVSep
	TokenDirSep
	TokenTupStart
	TokenTupEnd
	TokenTupSep
	TokenVarStart
	TokenVarEnd
	TokenVarSep
	TokenStrMark

	TokenWhitespace
	TokenNewLine
	TokenOther
	TokenEnd
)

var specialTokensByRune = map[rune]TokenKind{
	parser.KVSep:    TokenKVSep,
	parser.DirSep:   TokenDirSep,
	parser.TupStart: TokenTupStart,
	parser.TupEnd:   TokenTupEnd,
	parser.TupSep:   TokenTupSep,
	parser.VarStart: TokenVarStart,
	parser.VarEnd:   TokenVarEnd,
	parser.VarSep:   TokenVarSep,
	parser.StrMark:  TokenStrMark,
}

type state int

const (
	stateWhitespace state = iota
	stateNewline
	stateDirPart
	stateString
	stateOther
)

var primaryKindByState = map[state]TokenKind{
	stateWhitespace: TokenWhitespace,
	stateNewline:    TokenNewLine,
	stateDirPart:    TokenOther,
	stateString:     TokenOther,
	stateOther:      TokenOther,
}

type Scanner struct {
	reader *bufio.Reader
	token  strings.Builder
	state  state
}

func New(rd io.Reader) Scanner {
	return Scanner{reader: bufio.NewReader(rd)}
}

func (x *Scanner) Token() string {
	return x.token.String()
}

func (x *Scanner) Scan() (kind TokenKind, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				kind = TokenInvalid
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
				return TokenEnd, nil
			}
			return primaryKindByState[x.state], nil
		}

		if kind, ok := specialTokensByRune[r]; ok {
			if x.token.Len() > 0 {
				x.unread()
				return primaryKindByState[x.state], nil
			}

			switch r {
			case parser.DirSep:
				switch x.state {
				case stateString:
					break
				default:
					x.state = stateDirPart
				}

			case parser.StrMark:
				switch x.state {
				case stateString:
					x.state = stateWhitespace
				default:
					x.state = stateString
				}

			default:
				switch x.state {
				case stateString:
					break
				default:
					x.state = stateWhitespace
				}
			}

			x.append(r)
			return kind, nil
		}

		if strings.ContainsRune(whitespace, r) {
			switch x.state {
			case stateOther:
				x.unread()
				kind := primaryKindByState[x.state]
				x.state = stateWhitespace
				return kind, nil

			default:
				x.append(r)
				continue
			}
		}

		if strings.ContainsRune(newline, r) {
			switch x.state {
			case stateWhitespace:
				x.state = stateNewline
				x.append(r)
				continue

			case stateOther:
				x.unread()
				kind := primaryKindByState[x.state]
				x.state = stateNewline
				return kind, nil

			default:
				x.append(r)
				continue
			}
		}

		switch x.state {
		case stateWhitespace, stateNewline:
			if x.token.Len() == 0 {
				x.state = stateOther
				x.append(r)
				continue
			}
			x.unread()
			kind := primaryKindByState[x.state]
			x.state = stateOther
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
