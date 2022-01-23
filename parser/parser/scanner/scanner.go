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

var specials = map[rune]TokenKind{
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

var kindByState = map[state]TokenKind{
	stateWhitespace: TokenWhitespace,
	stateNewline:    TokenNewLine,
	stateDirPart:    TokenOther,
	stateString:     TokenOther,
	stateOther:      TokenOther,
}

func toTokenKind(in state) TokenKind {
	out, ok := kindByState[in]
	if !ok {
		panic(errors.Errorf("unexpected state %v", in))
	}
	return out
}

type Scanner struct {
	reader *bufio.Reader
	token  strings.Builder

	escape bool
	state  state
}

func New(rd io.Reader) Scanner {
	return Scanner{reader: bufio.NewReader(rd)}
}

func (x *Scanner) Token() string {
	return x.token.String()
}

func (x *Scanner) append(r rune) {
	_, err := x.token.WriteRune(r)
	if err != nil {
		panic(err)
	}
}

func (x *Scanner) unread() {
	err := x.reader.UnreadRune()
	if err != nil {
		panic(err)
	}
}

func (x *Scanner) Scan() (TokenKind, error) {
	x.token.Reset()

	for {
		r, _, err := x.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if x.token.Len() == 0 {
					return TokenEnd, nil
				}
				return toTokenKind(x.state), nil
			}
			return TokenInvalid, err
		}

		if kind, ok := specials[r]; ok {
			if x.token.Len() > 0 {
				x.unread()
				return toTokenKind(x.state), nil
			}

			switch kind {
			case TokenDirSep:
				switch x.state {
				case stateString:
					break
				default:
					x.state = stateDirPart
				}

			case TokenStrMark:
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

		switch x.state {
		case stateWhitespace:
			if strings.ContainsRune(whitespace, r) {
				x.append(r)
				continue
			}
			if strings.ContainsRune(newline, r) {
				x.state = stateNewline
				x.append(r)
				continue
			}
			if x.token.Len() == 0 {
				x.state = stateOther
				x.append(r)
				continue
			}
			x.unread()
			kind := toTokenKind(x.state)
			x.state = stateOther
			return kind, nil

		case stateNewline:
			if strings.ContainsRune(newline, r) {
				x.append(r)
				continue
			}
			if x.token.Len() == 0 {
				x.state = stateOther
				x.append(r)
				continue
			}
			x.unread()
			kind := toTokenKind(x.state)
			x.state = stateOther
			return kind, nil

		case stateDirPart:
			x.append(r)
			continue

		case stateString:
			x.append(r)
			continue

		case stateOther:
			if strings.ContainsRune(whitespace, r) {
				x.unread()
				kind := toTokenKind(x.state)
				x.state = stateWhitespace
				return kind, nil
			}
			if strings.ContainsRune(newline, r) {
				x.unread()
				kind := toTokenKind(x.state)
				x.state = stateNewline
				return kind, nil
			}
			x.append(r)
			continue
		}
	}
}
