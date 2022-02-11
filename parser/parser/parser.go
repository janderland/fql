package parser

import (
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type parserState int

const (
	parserStateInitial parserState = iota
	parserStateDirTail
	parserStateDirVarEnd
	parserStateDirHead
	parserStateTupleTail
	parserStateTupleHead
	parserStateSeparator
	parserStateValue
)

var parserStateName = map[parserState]string{
	parserStateDirTail:   "directory",
	parserStateDirHead:   "directory",
	parserStateTupleTail: "key's tuple",
	parserStateTupleHead: "key's tuple",
	parserStateSeparator: "query",
	parserStateValue:     "value",
}

var tokenKindName = map[TokenKind]string{
	TokenKindEscape:     "escape",
	TokenKindKVSep:      "key-value separator",
	TokenKindDirSep:     "directory separator",
	TokenKindTupStart:   "tuple start",
	TokenKindTupEnd:     "tuple end",
	TokenKindTupSep:     "tuple separator",
	TokenKindVarStart:   "variable start",
	TokenKindVarEnd:     "variable end",
	TokenKindVarSep:     "variable separator",
	TokenKindStrMark:    "string mark",
	TokenKindWhitespace: "whitespace",
	TokenKindNewLine:    "newline",
	TokenKindOther:      "other",
	TokenKindEnd:        "end of query",
}

type Token struct {
	Kind  TokenKind
	Token string
}

type Error struct {
	Tokens []Token
	Index  int
	Err    error
}

func (x *Error) Error() string {
	var msg strings.Builder
	for i, token := range x.Tokens {
		if i == x.Index {
			msg.WriteString(" <--- ")
		}
		msg.WriteString(token.Token)
	}
	return errors.Wrap(x.Err, msg.String()).Error()
}

type Parser struct {
	scanner Scanner
	tokens  []Token
	state   parserState
}

func NewParser(s Scanner) Parser {
	return Parser{scanner: s}
}

func (x *Parser) Parse() (q.Query, error) {
	var kv q.KeyValue

	for {
		kind, err := x.scanner.Scan()
		if err != nil {
			return nil, err
		}

		token := x.scanner.Token()
		x.tokens = append(x.tokens, Token{
			Kind:  kind,
			Token: token,
		})

		switch x.state {
		case parserStateInitial:
			switch kind {
			case TokenKindDirSep:
				x.state = parserStateDirHead

			default:
				return nil, x.withTokens(x.stateErr(kind))
			}

		case parserStateDirTail:
			switch kind {
			case TokenKindDirSep:
				x.state = parserStateDirHead

			case TokenKindTupStart:
				x.state = parserStateTupleHead

			case TokenKindEscape, TokenKindOther:
				if kind == TokenKindEscape {
					switch token[1] {
					case DirSep:
					default:
						return nil, x.withTokens(x.escapeErr(token))
					}
				}
				i := len(kv.Key.Directory) - 1
				str := kv.Key.Directory[i].(q.String)
				kv.Key.Directory[i] = q.String(string(str) + token)

			case TokenKindEnd:
				return kv.Key.Directory, nil

			default:
				return nil, x.withTokens(x.stateErr(kind))
			}

		case parserStateDirVarEnd:
			switch kind {
			case TokenKindVarEnd:
				kv.Key.Directory = append(kv.Key.Directory, q.Variable{})
				x.state = parserStateDirTail

			default:
				return nil, x.withTokens(x.stateErr(kind))
			}

		case parserStateDirHead:
			switch kind {
			case TokenKindVarStart:
				x.state = parserStateDirVarEnd

			case TokenKindEscape, TokenKindOther:
				x.state = parserStateDirTail
				kv.Key.Directory = append(kv.Key.Directory, q.String(token))

			default:
				return nil, x.withTokens(x.stateErr(kind))
			}
		}
	}
}

func (x *Parser) withTokens(err error) error {
	out := Error{
		Index: len(x.tokens),
		Err:   err,
	}

	for {
		kind, err := x.scanner.Scan()
		if err != nil {
			return err
		}

		if kind == TokenKindEnd {
			out.Tokens = x.tokens
			return &out
		}

		x.tokens = append(x.tokens, Token{
			Kind:  kind,
			Token: x.scanner.Token(),
		})
	}
}

func (x *Parser) escapeErr(token string) error {
	return errors.Errorf("unexpected escape '%v' while parsing %v", token, parserStateName[x.state])
}

func (x *Parser) stateErr(kind TokenKind) error {
	return errors.Errorf("unexpected %v while parsing %v", tokenKindName[kind], parserStateName[x.state])
}
