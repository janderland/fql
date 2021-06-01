package parser

import "strconv"

const (
	KVSep  = '='
	DirSep = '/'

	TupStart = '{'
	TupEnd   = '}'
	TupSep   = ','

	VarStart = '('
	VarEnd   = ')'
	VarSep   = '|'

	StrStart = '"'
	StrHex   = "\\x"
	StrEnd   = '"'

	Nil   = "nil"
	True  = "true"
	False = "false"
	Clear = "clear"
)

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
