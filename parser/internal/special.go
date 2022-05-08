package internal

const (
	// Whitespace contains the runes allowed to be
	// in a whitespace token.
	Whitespace = "\t "

	// Newline together with Whitespace contains the
	// runes allowed to be in a newline token.
	Newline = "\n\r"
)

// These are single-rune tokens. When these runes are
// encountered by the scanner.Scanner, they are usually
// returned in their own single-rune token string.
const (
	KeyValSep = '='
	DirSep    = '/'
	TupStart  = '{'
	TupSep    = ','
	TupEnd    = '}'
	VarStart  = '<'
	VarSep    = '|'
	VarEnd    = '>'
	StrMark   = '"'
)

const (
	// Escape marks the start of an escape token.
	Escape = '\\'

	// HexStart marks the start of a hexadecimal number token.
	HexStart = "0x"

	// Nil token string.
	Nil = "nil"

	// True token string.
	True = "true"

	// False token string.
	False = "false"

	// Clear token string.
	Clear = "clear"

	// MaybeMore token string.
	MaybeMore = "..."
)
