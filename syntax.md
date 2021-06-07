# FDBQ Language Syntax

This document formally defines the syntax of the FDBQ query
language. The meta language used is extended Backus-Naur
form as defined in ISO/IEC 14977.

Many rules contain references to `<ws>` which represents
contiguous whitespace. At the end of this document, there is
a postlude showing how this whitespace flexibility can be
useful.

```ebnf
ws = { ' ' | '\t' | '\n' | '\r' } ;
```

A `query` represents either a read, write, or clear
operation. The syntax structure mirrors the structure of an
FDB key-value. If the `query` contains a `variable` then it
is a read query (see `variable` for more information on read
queries). If the `query` contains `'clear'` as the `value`
then it's a clear query. Otherwise, it's a write query.

```ebnf
query = keyval | key | directory ;
```

A `keyval` query will result in reading key-values if it
contains a `variable`, clearing a key-value if it contains
`'clear'` (see the `value` rule), or setting a key-value if
it contains neither.

```ebnf
keyval = key, ws, '=', ws, value ;
```

Examples of a `keyval`:

```fdbq
/my/dir{33,0xFFE12A,c0cd12d7-8cb0-44bc-ad44-28af3f40c33e} = {"yup"}

/var/local{<int|float>,<>} = 8

{"hello","world"}=clear
```

A `key` query is equivalent to a `keyval` query where the
`value` is an empty `variable`: `'<>'`. For this reason, all
`key` queries result in read operations.

```enbf
key = directory, ws, tuple ;
```

Examples of a `key`:

```enbf
/my/dir{"welcome",33.9}

/this /location {"yelp","useless"}

/root/money { 33.9, "dollars" }
```

`value`

```
value = tuple | data | 'clear' ;
```

directory ::= '/' <ws> <directory> <ws> <directory> | '/' <directory>

tuple ::= '(' <ws> <elements> <ws> ')'

elements ::= <data> <ws> ',' <ws> <elements> | <tuple> <ws> ',' <ws> <elements> | <data> | <tuple>

data ::= 'nil' | <bool> | <int> | <float> | <scientific> | <string> | <uuid> | <base64>

bool ::= 'true' | 'false'

int ::= <number> | '-' <number>

float ::= <int> '.' <number>

scientific ::= <int> 'e' <int> | <float> 'e' <int>

string ::= '"' <words> '"'

uuid ::= 8 * <digit> '-' 4 * <digit> '-' 4 * <digit> '-' 4 * <digit> '-' 12 * <digit>

number ::= <digit> <number> | <digit>

words ::= <character> <words> | <character>

; <digit> is all single digit numbers 0-9.
; <character> is all ASCII characters.

