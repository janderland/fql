package parser

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/parser/internal"
	"github.com/janderland/fdbq/parser/parser/scanner"
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
	parserStateValueVarHead
	parserStateValueVarTail
	parserStateFinished
)

func parserStateName(state parserState) string {
	switch state {
	case parserStateInitial:
		return "initial"
	case parserStateDirHead:
		return "directory"
	case parserStateDirTail:
		return "directory"
	case parserStateDirVarEnd:
		return "directory"
	case parserStateTupleHead:
		return "tuple"
	case parserStateTupleTail:
		return "tuple"
	case parserStateTupleVarHead:
		return "variable"
	case parserStateTupleVarTail:
		return "variable"
	case parserStateTupleString:
		return "string"
	case parserStateSeparator:
		return "query"
	case parserStateValue:
		return "value"
	case parserStateValueVarHead:
		return "variable"
	case parserStateValueVarTail:
		return "variable"
	case parserStateFinished:
		return "finished"
	default:
		return fmt.Sprintf("[unknown parser state %v]", state)
	}
}

func tokenKindName(kind scanner.TokenKind) string {
	switch kind {
	case scanner.TokenKindEscape:
		return "escape"
	case scanner.TokenKindKVSep:
		return "key-value separator"
	case scanner.TokenKindDirSep:
		return "directory separator"
	case scanner.TokenKindTupStart:
		return "tuple start"
	case scanner.TokenKindTupEnd:
		return "tuple end"
	case scanner.TokenKindTupSep:
		return "tuple separator"
	case scanner.TokenKindVarStart:
		return "variable start"
	case scanner.TokenKindVarEnd:
		return "variable end"
	case scanner.TokenKindVarSep:
		return "variable separator"
	case scanner.TokenKindStrMark:
		return "string mark"
	case scanner.TokenKindWhitespace:
		return "whitespace"
	case scanner.TokenKindNewline:
		return "newline"
	case scanner.TokenKindOther:
		return "other"
	case scanner.TokenKindEnd:
		return "end of query"
	default:
		return fmt.Sprintf("[unknown token kind %v]", kind)
	}
}

type Token struct {
	Kind  scanner.TokenKind
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
	scanner scanner.Scanner
	tokens  []Token
	state   parserState
}

func NewParser(s scanner.Scanner) Parser {
	return Parser{scanner: s}
}

func (x *Parser) Parse() (q.Query, error) {
	var (
		kv  internal.KVBuilder
		tup internal.TupBuilder

		valTup bool
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
			case scanner.TokenKindDirSep:
				x.state = parserStateDirHead

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirTail:
			switch kind {
			case scanner.TokenKindDirSep:
				x.state = parserStateDirHead

			case scanner.TokenKindTupStart:
				x.state = parserStateTupleHead
				tup = internal.TupBuilder{}
				valTup = false

			case scanner.TokenKindEscape, scanner.TokenKindOther:
				if kind == scanner.TokenKindEscape {
					switch token[1] {
					case internal.DirSep:
					default:
						return nil, x.withTokens(x.escapeErr(token))
					}
				}
				kv.AppendToLastDirPart(token)

			case scanner.TokenKindEnd:
				return kv.Get().Key.Directory, nil

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirVarEnd:
			switch kind {
			case scanner.TokenKindVarEnd:
				x.state = parserStateDirTail
				kv.AppendVarToDirectory()

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateDirHead:
			switch kind {
			case scanner.TokenKindVarStart:
				x.state = parserStateDirVarEnd

			case scanner.TokenKindEscape, scanner.TokenKindOther:
				x.state = parserStateDirTail
				kv.AppendPartToDirectory(token)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleHead:
			switch kind {
			case scanner.TokenKindTupStart:
				tup.StartSubTuple()

			case scanner.TokenKindTupEnd:
				x.state = parserStateSeparator

			case scanner.TokenKindVarStart:
				x.state = parserStateTupleVarHead
				tup.Append(q.Variable{})

			case scanner.TokenKindStrMark:
				x.state = parserStateTupleString
				tup.Append(q.String(""))

			case scanner.TokenKindWhitespace, scanner.TokenKindNewline:
				break

			case scanner.TokenKindOther:
				x.state = parserStateTupleTail
				if token == internal.MaybeMore {
					tup.Append(q.MaybeMore{})
					break
				}
				data, err := parseData(token)
				if err != nil {
					return nil, x.withTokens(err)
				}
				tup.Append(data)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleTail:
			switch kind {
			case scanner.TokenKindTupEnd:
				if tup.EndTuple() {
					if valTup {
						x.state = parserStateFinished
						kv.SetValue(tup.Get())
						break
					}
					x.state = parserStateSeparator
					kv.SetKeyTuple(tup.Get())
				}

			case scanner.TokenKindTupSep:
				x.state = parserStateTupleHead

			case scanner.TokenKindWhitespace, scanner.TokenKindNewline:
				break

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleString:
			if kind == scanner.TokenKindEnd {
				return nil, x.withTokens(x.tokenErr(kind))
			}
			if kind == scanner.TokenKindStrMark {
				x.state = parserStateTupleTail
				break
			}
			tup.AppendToLastElemStr(token)

		case parserStateTupleVarHead:
			switch kind {
			case scanner.TokenKindVarEnd:
				x.state = parserStateTupleTail

			case scanner.TokenKindOther:
				x.state = parserStateTupleVarTail
				v, err := parseValueType(token)
				if err != nil {
					return nil, x.withTokens(err)
				}
				tup.AppendToLastElemVar(v)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateTupleVarTail:
			switch kind {
			case scanner.TokenKindVarEnd:
				x.state = parserStateTupleTail

			case scanner.TokenKindVarSep:
				x.state = parserStateTupleVarHead

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateSeparator:
			switch kind {
			case scanner.TokenKindEnd:
				return kv.Get().Key, nil

			case scanner.TokenKindKVSep:
				x.state = parserStateValue

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateValue:
			switch kind {
			case scanner.TokenKindTupStart:
				x.state = parserStateTupleHead
				tup = internal.TupBuilder{}
				valTup = true

			case scanner.TokenKindVarStart:
				x.state = parserStateValueVarHead
				kv.SetValue(q.Variable{})

			case scanner.TokenKindOther:
				x.state = parserStateFinished
				if token == internal.Clear {
					kv.SetValue(q.Clear{})
					break
				}
				data, err := parseData(token)
				if err != nil {
					return nil, x.withTokens(err)
				}
				kv.SetValue(data)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateValueVarHead:
			switch kind {
			case scanner.TokenKindVarEnd:
				x.state = parserStateFinished

			case scanner.TokenKindOther:
				x.state = parserStateValueVarTail
				v, err := parseValueType(token)
				if err != nil {
					return nil, x.withTokens(err)
				}
				kv.AppendToValueVar(v)

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateValueVarTail:
			switch kind {
			case scanner.TokenKindVarEnd:
				x.state = parserStateFinished

			case scanner.TokenKindVarSep:
				x.state = parserStateValueVarHead

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		case parserStateFinished:
			switch kind {
			case scanner.TokenKindWhitespace:
				break

			case scanner.TokenKindEnd:
				return kv.Get(), nil

			default:
				return nil, x.withTokens(x.tokenErr(kind))
			}

		default:
			return nil, errors.Errorf("unexpected state '%v'", parserStateName(x.state))
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

		if kind == scanner.TokenKindEnd {
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
	return errors.Errorf("unexpected escape '%v' while parsing %v", token, parserStateName(x.state))
}

func (x *Parser) tokenErr(kind scanner.TokenKind) error {
	return errors.Errorf("unexpected %v while parsing %v", tokenKindName(kind), parserStateName(x.state))
}

func parseValueType(token string) (q.ValueType, error) {
	for _, v := range q.AllTypes() {
		if string(v) == token {
			return v, nil
		}
	}
	return q.AnyType, errors.Errorf("unrecognized value type")
}

func parseData(token string) (
	interface {
		TupElement(q.TupleOperation)
		Value(q.ValueOperation)
		Eq(interface{}) bool
	},
	error,
) {
	if token == internal.Nil {
		return q.Nil{}, nil
	}
	if token == internal.True {
		return q.Bool(true), nil
	}
	if token == internal.False {
		return q.Bool(false), nil
	}
	if strings.HasPrefix(token, internal.HexStart) {
		token = token[len(internal.HexStart):]
		if len(token)%2 != 0 {
			return nil, errors.New("expected even number of hex digits")
		}
		data, err := hex.DecodeString(token)
		if err != nil {
			return nil, err
		}
		return q.Bytes(data), nil
	}
	if strings.Count(token, "-") == 4 {
		return parseUUID(token)
	}
	i, err := strconv.ParseInt(token, 10, 64)
	if err == nil {
		return q.Int(i), nil
	}
	u, err := strconv.ParseUint(token, 10, 64)
	if err == nil {
		return q.Uint(u), nil
	}
	f, err := strconv.ParseFloat(token, 64)
	if err == nil {
		return q.Float(f), nil
	}
	return nil, errors.New("unrecognized data element")
}

func parseUUID(token string) (q.UUID, error) {
	groups := strings.Split(token, "-")
	checkLen := func(i int, expLen int) error {
		if len(groups[i]) != expLen {
			return errors.Errorf("the %s group should contain %d characters rather than %d", ordinal(i+1), expLen, len(groups[i]))
		}
		return nil
	}
	if err := checkLen(0, 8); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(1, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(2, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(3, 4); err != nil {
		return q.UUID{}, err
	}
	if err := checkLen(4, 12); err != nil {
		return q.UUID{}, err
	}

	var uuid q.UUID
	_, err := hex.Decode(uuid[:], []byte(strings.ReplaceAll(token, "-", "")))
	if err != nil {
		return q.UUID{}, err
	}
	return uuid, nil
}

func ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return strconv.Itoa(x) + suffix
}
