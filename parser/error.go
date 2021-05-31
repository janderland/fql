package parser

import (
	"strings"
)

var _ error = Error{}

type Error struct {
	kind  string
	token string
	cause error
}

func newErrWrapper(kind, token string) func(error) error {
	return func(cause error) error {
		if cause == nil {
			return nil
		}
		return Error{
			kind:  kind,
			token: token,
			cause: cause,
		}
	}
}

func (e Error) Error() string {
	var s strings.Builder
	e.error(&s)
	return s.String()
}

func (e Error) error(s *strings.Builder) {
	s.WriteString("failed to parse ")
	s.WriteString(e.kind)
	s.WriteString(" - '")
	s.WriteString(e.token)
	s.WriteRune('\'')

	if e.cause == nil {
		return
	}

	s.WriteString(": ")
	switch cause := e.cause.(type) {
	case Error:
		cause.error(s)
	default:
		s.WriteString(cause.Error())
	}
}

func (e Error) SPrint() string {
	var s strings.Builder
	s.WriteString("failed to parse!\n")
	e.sprint(&s)
	return s.String()
}

func (e Error) sprint(s *strings.Builder) {
	s.WriteString("  ")
	s.WriteString(e.kind)
	s.WriteString(": '")
	s.WriteString(e.token)
	s.WriteString("'\n")

	if e.cause == nil {
		return
	}

	switch cause := e.cause.(type) {
	case Error:
		cause.sprint(s)
	default:
		s.WriteString("  ")
		s.WriteString(cause.Error())
	}
}
