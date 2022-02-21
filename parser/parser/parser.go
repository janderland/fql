package parser

import (
	"strconv"
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type parserState int

const (
	parserStateInitial parserState = iota
	parserStateDirHead
	parserStateDirTail
	parserStateDirVarEnd
	parserStateTupleHead
	parserStateTupleTail
	parserStateTupleVarHead
	parserStateTupleVarTail
	parserStateTupleString
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
		if i+1 == x.Index {
			msg.WriteString(" --> ")
		}
		msg.WriteString(token.Token)
		if i+1 == x.Index {
			msg.WriteString(" <--invalid-token--- ")
		}
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
	var (
		kv  kvBuilder
		tup tupBuilder
	)

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
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirTail:
			switch kind {
			case TokenKindDirSep:
				x.state = parserStateDirHead

			case TokenKindTupStart:
				x.state = parserStateTupleHead
				tup = tupBuilder{}

			case TokenKindEscape, TokenKindOther:
				if kind == TokenKindEscape {
					switch token[1] {
					case DirSep:
					default:
						return nil, x.withTokens(x.escapeErr(token))
					}
				}
				kv.appendToLastDirPart(token)

			case TokenKindEnd:
				return kv.get().Key.Directory, nil

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirVarEnd:
			switch kind {
			case TokenKindVarEnd:
				x.state = parserStateDirTail
				kv.appendVarToDirectory()

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirHead:
			switch kind {
			case TokenKindVarStart:
				x.state = parserStateDirVarEnd

			case TokenKindEscape, TokenKindOther:
				x.state = parserStateDirTail
				kv.appendPartToDirectory(token)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleHead:
			switch kind {
			case TokenKindTupStart:
				tup.startSubTuple()

			case TokenKindTupEnd:
				x.state = parserStateSeparator

			case TokenKindVarStart:
				x.state = parserStateTupleVarHead

			case TokenKindStrMark:
				x.state = parserStateTupleString
				tup.appendToTuple(q.String(""))

			case TokenKindWhitespace, TokenKindNewLine:
				break

			case TokenKindOther:
				x.state = parserStateTupleTail
				var num q.TupElement
				i, err := strconv.ParseInt(token, 10, 64)
				if num == nil && err == nil {
					num = q.Int(i)
				}
				u, err := strconv.ParseUint(token, 10, 64)
				if num == nil && err == nil {
					num = q.Uint(u)
				}
				f, err := strconv.ParseFloat(token, 64)
				if num == nil && err == nil {
					num = q.Float(f)
				}
				if num != nil {
					tup.appendToTuple(num)
					break
				}
				return nil, x.withTokens(errors.Errorf("invalid tuple element"))

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleTail:
			switch kind {
			case TokenKindTupEnd:
				if tup.endCurrentTuple() {
					x.state = parserStateSeparator
					kv.setKeyTuple(tup.get())
				}

			case TokenKindTupSep:
				x.state = parserStateTupleHead

			case TokenKindWhitespace, TokenKindNewLine:
				break

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleString:
			if kind == TokenKindEnd {
				return nil, x.withTokens(x.tokenErr(kind))
			}
			if kind == TokenKindStrMark {
				x.state = parserStateTupleTail
				break
			}
			tup.appendToLastTupElem(token)

		case parserStateSeparator:
			switch kind {
			case TokenKindEnd:
				return kv.get().Key, nil
			}

		default:
			return nil, errors.Errorf("unexpected state %v", parserStateName[x.state])
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

func (x *Parser) tokenErr(kind TokenKind) error {
	return errors.Errorf("unexpected %v while parsing %v", tokenKindName[kind], parserStateName[x.state])
}
