package parser

import (
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
	return x.Err.Error()
}

func Parse(scanner Scanner) (q.Query, error) {
	var (
		tokens []Token
		state  parserState
		kv     q.KeyValue
	)

	withTokens := func(err error) error {
		out := Error{
			Index: len(tokens),
			Err:   err,
		}

		for {
			kind, err := scanner.Scan()
			if err != nil {
				return err
			}

			if kind == TokenKindEnd {
				out.Tokens = tokens
				return &out
			}

			tokens = append(tokens, Token{Kind: kind, Token: scanner.Token()})
		}
	}

	for {
		kind, err := scanner.Scan()
		if err != nil {
			return nil, err
		}

		token := scanner.Token()
		tokens = append(tokens, Token{Kind: kind, Token: token})

		switch state {
		case parserStateInitial:
			switch kind {

			default:
				return nil, withTokens(stateErr(kind, state))
			}

		case parserStateDirTail:
			switch kind {
			case TokenKindDirSep:
				state = parserStateDirHead

			case TokenKindTupStart:
				state = parserStateTupleHead

			case TokenKindEscape, TokenKindOther:
				if kind == TokenKindEscape {
					switch token[1] {
					case DirSep:
					default:
						return nil, withTokens(escapeErr(token, state))
					}
				}
				i := len(kv.Key.Directory) - 1
				str := kv.Key.Directory[i].(q.String)
				kv.Key.Directory[i] = q.String(string(str) + token)

			case TokenKindEnd:
				return kv.Key.Directory, nil

			default:
				return nil, withTokens(stateErr(kind, state))
			}

		case parserStateDirVarEnd:
			switch kind {
			case TokenKindVarEnd:
				kv.Key.Directory = append(kv.Key.Directory, q.Variable{})
				state = parserStateDirTail

			default:
				return nil, withTokens(stateErr(kind, state))
			}

		case parserStateDirHead:
			switch kind {
			case TokenKindVarStart:
				state = parserStateDirVarEnd

			case TokenKindEscape, TokenKindOther:
				state = parserStateDirTail
				kv.Key.Directory = append(kv.Key.Directory, q.String(token))

			case TokenKindEnd:
				return kv.Key.Directory, nil

			default:
				return nil, withTokens(stateErr(kind, state))
			}
		}
	}
}

func escapeErr(token string, state parserState) error {
	return errors.Errorf("unexpected escape '%v' while parsing %v", token, parserStateName[state])
}

func stateErr(kind TokenKind, state parserState) error {
	return errors.Errorf("unexpected %v while parsing %v", tokenKindName[kind], parserStateName[state])
}
