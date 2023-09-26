package wrap

import (
	rw "github.com/mattn/go-runewidth"
	"strings"
	"unicode"
)

func Wrap(str string, limit int) []string {
	if limit == 0 {
		return strings.Split(str, "\n")
	}

	var (
		word     builder
		line     builder
		lines    []string
		ansiCode = false
	)

	for _, c := range str {
		switch {
		// TODO: Double check the ASCII escape codes.
		case c == '\x1B':
			word.WriteCode(c)
			ansiCode = true

		case ansiCode:
			word.WriteCode(c)
			if (c >= 0x40 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a) {
				ansiCode = false
			}

		case unicode.IsSpace(c):
			line.WriteString(word.String())
			word.Reset()

			if line.Width() < limit {
				line.WriteRune(' ')
				continue
			}

			lines = append(lines, line.String())
			line.Reset()

		default:
			// Does the char fit into the word?
			// Well, if we add the char into the
			// word does the word still fit into
			// the line?
			if line.Width()+word.Width()+rw.RuneWidth(c) > limit {
				line.WriteString(word.String())
				word.Reset()

				lines = append(lines, line.String())
				line.Reset()
			}

			word.WriteRune(c)
		}
	}

	line.WriteString(word.String())
	lines = append(lines, line.String())

	return lines
}

type builder struct {
	builder strings.Builder
	width   int
}

func (x *builder) Width() int {
	return x.width
}

func (x *builder) Reset() {
	x.width = 0
	x.builder.Reset()
}

func (x *builder) String() string {
	return x.builder.String()
}

func (x *builder) WriteCode(r rune) {
	_, _ = x.builder.WriteRune(r)
}

func (x *builder) WriteRune(r rune) {
	x.width += rw.RuneWidth(r)
	_, _ = x.builder.WriteRune(r)
}

func (x *builder) WriteString(s string) {
	for _, c := range s {
		x.width += rw.RuneWidth(c)
	}
	_, _ = x.builder.WriteString(s)
}
