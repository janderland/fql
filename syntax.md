; TODO: Use ABNF - https://datatracker.ietf.org/doc/html/rfc5234

query ::= <keyval> | <key> | <dirpath>

keyval ::= <key> <ws> '=' <ws> <value>

key ::= <path> <ws> <tuple>

value ::= <tuple> | <data> | 'clear'

dirpath ::= '/' <ws> <directory> <ws> <dirpath> | '/' <directory>

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

ws ::= ' ' <ws> | ' ' | ''

; <digit> is all single digit numbers 0-9.
; <character> is all ASCII characters.

