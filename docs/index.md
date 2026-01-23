---
title: FQL

# We include this intro via the 'include-before'
# metadata field so it's placed before the TOC.
include-before: |
  ```language-fql {.query}
  /user/index/surname("Johnson",<userID:int>)
  /user(:userID,...)
  ```
  ```language-fql {.result}
  /user(9323,"Timothy","Johnson",37)=nil
  /user(24335,"Andrew","Johnson",42)=nil
  /user(33423,"Ryan","Johnson",0x0ffa83,42.2)=nil
  ```
  FQL is an [open source](https://github.com/janderland/fql)
  query language for
  [FoundationDB](https://www.foundationdb.org/). It's query
  semantics mirror FoundationDB's [core data
  model](https://apple.github.io/foundationdb/data-modeling.html).
  Fundamental patterns like range-reads and indirection are first
  class citizens.
...

# Introduction

This document serves as both a language specification and
a usage guide for FQL. The [Syntax](#syntax) section
describes the structure of queries while the
[Semantics](#semantics) section describes their behavior.
The [Implementations](#implementations) section highlights
features which are not included in FQL but may be defined by
a particular implementation. The complete [EBNF
grammar](#grammar) appears at the end.

Throughout the document, relevant grammar rules are shown
alongside the features they define. Python code snippets are
also included demonstrating equivalent FoundationDB API
calls.

Grammar rules use extended Backus-Naur form as defined in
ISO/IEC 14977, with a modification: concatenation and rule
termination are implicit.

> ❗ Not all features described in this document have been
> implemented yet. Callouts like this one will mark features
> that are still being worked on.

# Syntax

## Overview

FQL is specified as a context-free [grammar](#grammar). The
queries look like key-values encoded using the [directory][]
and [tuple][] layers. To the left of the `=` is the key
which includes a directory path and tuple. To the right is
the value.

[directory]: https://apple.github.io/foundationdb/developer-guide.html#directories
[tuple]: https://apple.github.io/foundationdb/data-modeling.html#data-modeling-tuples

```language-ebnf {.grammar}
query = keyval | key | dquery
dquery = directory [ '=' 'remove' ]
keyval = key '=' value
key = directory tuple
value = 'clear' | data
```

> The grammar above is slightly simplified.
> [Options](#options) should be included in the `query`
> rule, but will be saved till later for simplicity.

A query may be a full key-value, just a key, or just
a directory query.

```language-fql {.query}
/my/directory("my","tuple")=4000
```

FQL queries may define a single key-value to be written, as
shown above, or may define a set of key-values to be read,
as shown below.

```language-fql {.query}
/my/directory("my","tuple")=<int>
```

```language-fql {.result}
/my/directory("my","tuple")=4000
```

The query above has the variable `<int>` as its value.
Variables act as placeholders for any of the supported [data
elements](#data-elements). 

FQL queries may also perform [range reads][] and filtering
by including on or more variables in the key. The query
below will return all key-values which conform to the schema
defined by the query.

[range reads]: https://apple.github.io/foundationdb/developer-guide.html#range-reads

```language-fql {.query}
/my/directory(<>,"tuple")=nil
```

```language-fql {.result}
/my/directory("your","tuple")=nil
/my/directory(42,"tuple")=nil
```

The variable `<>` in the query above lacks a type. This
means the schema allows any [data element](#data-elements)
at the variable's position.

All key-values with a certain key prefix may be range read
by ending the tuple with `...`.

```language-fql {.query}
/my/directory("my","tuple",...)=<>
```

```language-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

A query's value may be omitted to imply the variable `<>`,
meaning the following query is semantically identical to the
one above.

```language-fql {.query}
/my/directory("my","tuple",...)
```

```language-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

Including a variable in the directory path tells FQL to
perform the read on all directory paths matching the schema.

```language-fql {.query}
/<>/directory("my","tuple")
```

```language-fql {.result}
/my/directory("my","tuple")=0x0fa0
/your/directory("my","tuple")=nil
```

Key-values may be cleared by using the special `clear` token
as the value.

```language-fql {.query}
/my/directory("my","tuple")=clear
```

The directory layer may be queried by only including
a directory path.

```language-fql {.query}
/my/<>
```

```language-fql {.result}
/my/directory
```

Directories are not explicitly created. During a write
query, the directory is created if it doesn't exist.
Directories may be removed by suffixing the directory path
with `=remove`.

```language-fql {.query} 
/my/dir=remove
```

## Data Elements

An FQL query contains instances of data elements. These
mirror the types of elements found in the [tuple layer][].
This section will describe how data elements behave in the
FQL language, while [element encoding](#data-encoding)
describes how FQL encodes the elements before writing them
to the DB.

[tuple layer]: https://github.com/apple/foundationdb/blob/main/design/tuple.md

<div>

| Type     | Description      | Examples                               |
|:---------|:-----------------|:---------------------------------------|
| `nil`    | Empty Type       | `nil`                                  |
| `bool`   | Boolean          | `true` `false`                         |
| `int`    | Signed Integer   | `-14` `3033`                           |
| `num`    | Floating Point   | `33.4` `-3.2e5`                        |
| `str`    | Unicode String   | `"happy😁"` `"\"quoted\""`             |
| `uuid`   | UUID             | `5a5ebefd-2193-47e2-8def-f464fc698e31` |
| `bytes`  | Byte String      | `0xa2bff2438312aac032`                 |
| `tup`    | Tuple            | `("hello",27.4,nil)`                   |
| `vstamp` | Version Stamp    | `#:0000` `#0102030405060708090a:0000`  |

</div>

The `nil` type may only be instantiated as the element
`nil`. The `bool` type may be instantiated as `true` or
`false`.

```language-ebnf {.grammar}
bool = 'true' | 'false'
```

The `int` type may be instantiated as any arbitrarily large
integer.

```language-ebnf {.grammar}
int = [ '-' ] digits
digits = digit { digit }
digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
```

The `num` type may be instantiated as any real number which
can be approximated by an [80-bit floating point][] value,
in accordance with IEEE 754. The implementation determines
the exact range of allowed values. Scientific notation may
be used. As expressed in the above specification, the type
may be instantiated as `-inf`, `inf`, `-nan` or `nan`.

[80-bit floating point]: https://en.wikipedia.org/wiki/Extended_precision#x86_extended_precision_format

```language-ebnf {.grammar}
num = int '.' digits | ( int | int '.' digits ) 'e' int | '-inf' | 'inf' | '-nan' | 'nan'
```

The `str` type may be instantiated as a unicode string
wrapped in double quotes. Quoted strings may contain double
quotes and backslashes via backslash escapes.

```language-ebnf {.grammar}
string = '"' { char | '\\"' | '\\\\' } '"'
```

The `uuid` and `bytes` types may be instantiated using
upper, lower, or mixed case hexidecimal numbers. For `uuid`,
the numbers are grouped in the standard 8, 4, 4, 4, 12
format. For `bytes`, any even number of hexidecimal digits
are prefixed by `0x`.

```language-ebnf {.grammar}
uuid = hex{8} '-' hex{4} '-' hex{4} '-' hex{4} '-' hex{12}
bytes = '0x' { hex hex }
hex = digit | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F'
```

The `tup` type may contain any of the data elements,
including nested tuples. Elements are separated by commas
and wrapped in parentheses. A trailing comma is allowed
after the last element. The last element may be the `...`
token (see [holes](#holes-&-schemas)).

```language-ebnf {.grammar}
tuple = '(' [ nl elements [ ',' ] nl ] ')'
elements = data [ ',' nl elements ] | '...'
```

The `vstamp` type represents a FoundationDB [versionstamp][].
A versionstamp contains a 10-byte transaction version and
a 2-byte user version. The transaction version is assigned
by the database at commit time. A vstamp without the
transaction version (only the user version after the colon)
is incomplete and will be filled in by FoundationDB when
written.

[versionstamp]: https://apple.github.io/foundationdb/api-general.html#versionstamps

```language-ebnf {.grammar}
vstamp = '#' [ hex{20} ] ':' hex{4}
```

## Directories

Directories provide a way to organize key-values into
hierarchical namespaces. The [directory layer][] manages
these namespaces and assigns short prefixes to keys.
Strings are the only element type allowed in directories.

[directory layer]: https://apple.github.io/foundationdb/developer-guide.html#directories

```language-ebnf {.grammar}
directory = '/' element [ directory ]
element = '<>' | name | string
name = { alphanumeric | '.' | '-' | '_' }
```

A directory is specified as a sequence of strings, each
prefixed by a forward slash. Strings do not need quotes if
they only contain alphanumericals, underscores, dashes, or
periods. To use other symbols, the strings must be quoted.

```language-fql {.query}
/my/dir/path_way
/my/"dir@--\o/"/path_way
```

The empty variable `<>` may be used in a directory path as
a placeholder for any directory name.

```language-fql {.query}
/root/<>/items
```

```language-fql {.result}
/root/good/items
/root/bad/items
/root/weird/items
```

## Holes & Schemas

Holes are a group of syntax constructs used to define
a key-value schema by acting as placeholders for one or more
data elements. There are three kinds of holes: variables,
references, and the `...` token.

> While the `...` token is semantically a hole, it is
> defined as part of the tuple's rule to restrict it's
> placement.

```language-ebnf {.grammar}
hole = variable | reference
variable = '<' [ name ':' ] [ type { '|' type } ] '>'
reference = ':' name
type = 'any' | 'tuple' | 'bool' | 'int' | 'num' | 'str' | 'uuid' | 'bytes' | 'vstamp'
```

Variables are used to represent a single [data
element](#data-elements). Variables are specified as a list
of element types, separated by `|`, wrapped in angled
braces.

```language-fql
<int|str|uuid|bytes>
```

The variable's type list describes which kinds of data
elements are allowed at the variable's position. A variable
may be empty, including no element types, meaning it allows
any element type.

```language-fql {.query}
/data(<int>,<str|int>,<>)=<>
```

```language-fql {.result}
/data(0,"jon",0xffab0c)=nil
/data(20,3,22.3)=0xff
/data(21,"",nil)=nil
```

References allow two queries to be connected via
a variable's name, allowing for [index
indirection](#indirection). Before the type list, a variable
may include a name. The reference is specified as
a variable's name prefixed with a `:`.

```language-fql {.query}
/index("car IDs",<varName:int>)
/data(:varName,...)
```

```language-fql {.result}
/data(33,"mazda")=nil
/data(320,"ford")=nil
/data(411,"chevy")=nil
```

Named variables must include at least one type. To allow
named variables to match any element type, use the `any`
type.

```language-fql
/stuff(<thing:any>)
```

```language-fql {.result}
/stuff("cat")
/stuff(42)
/stuff(0x5fae)
```

The `...` token represents any number of data elements of
any type. It is only allowed as the last element of a tuple.

```language-fql
/tuples(0x00,...)
```

```language-fql {.result}
/tuples(0x00)=nil
/tuples(0x00,"something")=nil
/tuples(0x00,42,43,44)=0xabcf
```

## Space & Comments

Whitespace and newlines are allowed within a tuple, between
its elements.

```language-fql {.query}
/account/private(
  <int>,
  <int>,
  <str>,
)=<int>
```

Comments start with a `%` and continue until the end of the
line. They can be used to document a tuple's elements.

```language-fql
% private account balances
/account/private(
  <int>,  % group ID
  <int>,  % account ID
  <str>,  % account name
)=<int>   % balance in USD
```

## Options

> ⚠️ Options are not implemented yet.

Options modify the semantics of [data
elements](#data-elements), [variables](#holes-schemas), and
[queries](#query-types). They can instruct FQL to use
alternative encodings, limit a query's result count, or
change other behaviors.

```language-ebnf {.grammar}
options = '[' option { ',' option } ']'
option = name [ ':' argument ]
argument = name | int | string
```

Options are specified as a comma separated list wrapped in
brackets. For instance, to specify that an `int` should be
encoded as a little-endian unsigned 8-bit integer, the
following options would be included after the element.

```language-fql
3548[u8,le]
```

Similarly, if a variable should only match against
a big-endian 32-bit floats then the following options would
be included after the `num` type.

```language-fql
<num[f32,be]>
```

### Element Options

The tables below show which options are supported for the
`int` and `num` types when used as values. These options
control how the data is serialized and unserialized.

<div>

| Int Option | Description     |
|:-----------|:----------------|
| `be`       | Big endian      |
| `le`       | Little endian   |
| `u8`       | Unsigned 8-bit  |
| `u16`      | Unsigned 16-bit |
| `u32`      | Unsigned 32-bit |
| `u64`      | Unsigned 64-bit |
| `i8`       | Signed 8-bit    |
| `i16`      | Signed 16-bit   |
| `i32`      | Signed 32-bit   |
| `i64`      | Signed 64-bit   |

</div>
<div>

| Num Option | Description   |
|:-----------|:--------------|
| `be`       | Big endian    |
| `le`       | Little endian |
| `f32`      | 32-bit        |
| `f64`      | 64-bit        |
| `f80`      | 80-bit        |

</div>

### Query Options

Query options are specified on the line before the query.
For instance, to specify that a range-read query should read
in reverse and only read 5 items, the following options
would be included before the query.

```language-fql {.query}
[reverse,limit:5]
/my/integers(<int>)=nil
```

Notice that the `limit` option includes a number after the
colon. Some options include a single argument to further
specify the option's behavior.

<div>

| Query Option | Argument | Description                        |
|:-------------|:---------|:-----------------------------------|
| `reverse`    | none     | Range read in reverse order        |
| `limit`      | int      | Maximum number of results          |
| `mode`       | name     | Range read mode: want_all, iterator, exact, small, medium, large, serial |
| `snapshot`   | none     | Use snapshot read                  |
| `strict`     | none     | Error when a read key-values doesn't conform to the schema |

</div>

# Semantics

## Data Encoding

FoundationDB stores the keys and values as simple byte
strings leaving the client responsible for encoding the
data. FQL determines how to encode [data
elements](#data-elements) based on their data type, position
within the query, and associated [options](#options).

Keys are *always* encoded using the [directory][] and
[tuple][] layers. Write queries create directories if they
do not exist.

```language-fql {.query}
/directory/"p@th"(nil,57223,0xa8ff03)=nil
```

```language-python {.equiv-py}
@fdb.transactional
def write_kv(tr):
    # Open directory; create if doesn't exist
    dir = fdb.directory.create_or_open(tr, ('directory', 'p@th'))

    # Pack the tuple and prepend the directory prefix
    key = dir.pack((None, 57223, b'\xa8\xff\x03'))

    # Write the KV
    tr[key] = b''
```

If a query reads from a directory which doesn't exist,
nothing is returned. The tuple layer encodes metadata about
element types, allowing FQL to decode keys without a schema.

```language-fql {.query}
/directory/<>(...)
```

```language-python {.equiv-py}
@fdb.transactional
def read_kvs(tr):
    # Open directory; exit if it doesn't exist
    dir = fdb.directory.open(tr, ('directory',))
    if dir is None:
        return []

    # List the sub-directories
    sub_dirs = dir.list(tr)

    # For each sub-directory, grab all the KVs
    results = []
    for sub_name in sub_dirs:
        sub_dir = dir.open(tr, (sub_name,))
        for key, val in tr[sub_dir.range()]:
            # Remove the directory prefix and unpack the tuple
            tup = sub_dir.unpack(key)
            # Value unpacking will be discussed later...
            results.append((sub_dir.get_path(), tup, val))

    return results
```

Values have more flexible encoding options. There is
a default encoding where data elements are encoded as the
lone member of a tuple. For instance, the value `42` is
encoded as the tuple `(42)`.

The exceptions to this default encoding are when values are
tuples (which are not wrapped in another tuple) and byte
strings (which are used as-is for the value).

```language-fql {.query}
/people/age("jon","smith")=42
```

```language-python {.equiv-py}
@fdb.transactional
def write_age(tr):
    dir = fdb.directory.create_or_open(tr, ('people', 'age'))
    key = dir.pack(('jon', 'smith'))

    # Pack the value as a tuple
    val = fdb.tuple.pack((42,))

    # Write the KV
    tr[key] = val
```

This default encoding allows values to be decoded without
knowing their type.

```language-fql {.query}
/people/age("jon","smith")=<>
```

```language-python {.equiv-py}
@fdb.transactional
def read_age(tr):
    dir = fdb.directory.open(tr, ('people', 'age'))
    key = dir.pack(('jon', 'smith'))

    # Read the value
    val_bytes = tr[key]

    # Assume the value is a tuple
    try:
        val_tup = fdb.tuple.unpack(val_bytes)
        if len(val_tup) == 1:
            return val_tup[0]
        return val_tup
    except:
        # If decoding as a tuple fails, return raw bytes
        return val_bytes
```

Using options, values can be encoded in other ways. For
instance, the option `u16` tells FQL to encode an integer as
an unsigned 16-bit integer. The byte order can be specified
using the options `le` and `be` for little and big endian
respectively. 

```language-fql {.query}
/numbers/big("37")=37[i16,be]
```

```language-python {.equiv-py}
import struct

@fdb.transactional
def write_int(tr):
    dir = fdb.directory.create_or_open(tr, ('numbers', 'big'))
    key = dir.pack(('37',))

    # Pack the value into signed 16-bit big endian
    val = struct.pack('>h', 37)

    # Write the KV
    tr[key] = val
```

If the value was encoded with non-default values, then the
encoding must be specified in the variable when read.
Otherwise, the default decoding will fail and it will be
returned as raw bytes.

```language-fql {.query}
/numbers/big("37")=<int[i16,be]>
```

```language-python {.equiv-py}
import struct

@fdb.transactional
def read_int(tr):
    dir = fdb.directory.open(tr, ('numbers', 'big'))
    key = dir.pack(('37',))

    # Read the value
    val_bytes = tr[key]

    # Unpack value as a 16-bit signed int, big endian
    val = struct.unpack('>h', val_bytes)[0]

    return val
```

## Query Types

FQL queries may write a single key-value, read/clear one or
more key-values, or list/remove directories. Although all
queries resemble key-values, their tokens imply one of the
above operations.

### Reads & Writes

Queries lacking [holes](#holes-schemas) perform writes on
the database. You can think of these queries as declaring
the existence of a particular key-value. Most query results
can be fed back into FQL as write queries. The exception to
this rule are [aggregate queries](#aggregation) and results
created by non-default [formatting](#formatting).

> ❗ Queries lacking a value altogether imply an empty
> [variable](#holes-schemas) as the value and should not be
> confused with write queries.

Queries containing [holes](#holes-schemas) read one or more
key-values. If the holes only appear in the value, then
a single key-value is returned, if one matching the schema
exists.

FQL attempts to decode the value as each of the types listed
in the variable, stopping at first success. If the value
cannot be decoded, the key-value does not match the schema.

Queries with [variables](#holes-schemas) in their key (and
optionally in their value) result in a range of key-values
being read.

Whether reading single or many, when a key-value is
encountered which doesn't match the query's schema it is
filtered out of the results. Including the `strict` [query
option](#query-options) causes the query to fail when
encountering a non-conformant key-value.

If a query has the token `clear` as it's value, it clears
all the key matching the query's schema. Keys not matching
the schema are ignored unless the `strict` option is
present, resulting in the query failing.

### Directories

The directory layer may be queried in isolation by using
a lone directory as a query. Directory queries are read-only
except when removing a directory. If the directory path
contains no variables, the query will read that single
directory.

A directory can be removed by appending `=remove` to the
directory query. If multiple directories match the schema,
they will all be removed.

### Filtering

Read queries define a schema to which key-values may or
may-not conform. In the Python snippets above, non-conformant
key-values were being filtered out of the results.

Alternatively, FQL can throw an error when encountering
non-conformant key-values. This may help enforce the
assumption that all key-values within a directory conform to
a certain schema. See the `strict` [query option](#query-options).

Because filtering is performed on the client side, range
reads may stream a lot of data to the client while the
client filters most of it away. For example, consider the
following query:

```language-fql {.query}
/people(3392,<str|int>,<>)=(<int>,...)
```

In the key, the location of the first variable or `...`
token determines the range read prefix used by FQL. For this
particular query, the prefix would be as follows:

```language-fql {.query}
/people(3392)
```

FoundationDB will stream all key-values with this prefix to
the client. As they are received, the client will filter out
key-values which don't match the query's schema. Below you
can see a Python implementation of how this filtering would
work.

```language-python
@fdb.transactional
def filter_range(tr):
    dir = fdb.directory.open(tr, ('people',))
    if dir is None:
        return []

    prefix = dir.pack((3392,))
    range_result = tr[fdb.Range(prefix, fdb.strinc(prefix))]

    results = []
    for key, val in range_result:
        tup = dir.unpack(key)

        # Our query specifies a key-tuple with 3 elements
        if len(tup) != 3:
            continue

        # The 2nd element must be either a string or an int
        if not isinstance(tup[1], (str, int)):
            continue

        # The query tells us to assume the value is a packed tuple
        try:
            val_tup = fdb.tuple.unpack(val)
        except:
            continue

        # The value-tuple must have one or more elements
        if len(val_tup) == 0:
            continue

        # The first element of the value-tuple must be an int
        if not isinstance(val_tup[0], int):
            continue

        results.append((tup, val_tup))

    return results
```

Filtering can also be combined with clearing. A filter clear
operation clears only the key-values that match the schema.

```language-fql {.query}
/people(3392,<str>,<>)=clear
```

This query clears all key-values under `/people` with prefix
`(3392)` where the second element is a string.

## Advanced Queries

Besides basic CRUD operations, FQL is capable of performing
indirection queries.

### Indirection

> ⚠️ Indirection is not implemented yet.

Indirection queries are similar to SQL joins. They associate
different groups of key-values via some shared data element.

In FoundationDB, indexes are implemented using indirection.
Suppose we have a large list of people, one key-value for
each person.

```language-fql {.query}
/people(
  <int>, % ID
  <str>, % First Name
  <str>, % Last Name
  <int>, % Age
)=nil
```

If we wanted to read all records containing the last name
"Johnson", we'd have to perform a linear search across the
entire "people" directory. To make this kind of search more
efficient, we can store an index for last names in
a separate directory.

```language-fql {.query}
/index/last_name(
  <str>, % Last Name
  <int>, % ID
)=nil
```

If we query the index, we can get the IDs of the records
containing the last name "Johnson".

```language-fql {.query}
/index/last_name("Johnson",<int>)
```

```language-fql {.result}
/index/last_name("Johnson",23)=nil
/index/last_name("Johnson",348)=nil
/index/last_name("Johnson",2003)=nil
```

FQL can forward the observed values of named variables from
one query to the next. We can use this to obtain our desired
subset from the "people" directory.

```language-fql {.query}
/index/last_name("Johnson",<id:int>)
/people(:id,...)
```

```language-fql {.result}
/people(23,"Lenny","Johnson",22,"Mechanic")=nil
/people(348,"Roger","Johnson",54,"Engineer")=nil
/people(2003,"Larry","Johnson",8,"N/A")=nil
```

### Aggregation

> ⚠️ Aggregation is not implemented yet.

Aggregation queries combine values from multiple key-values
into a single result. FQL provides pseudo data types which
perform aggregation, similar to SQL aggregate functions.

| Pseudo Type | Description                              |
|:------------|:-----------------------------------------|
| `count`     | Count the number of matching key-values  |
| `sum`       | Sum integer values                       |
| `avg`       | Average of integer values                |
| `min`       | Minimum value                            |
| `max`       | Maximum value                            |
| `append`    | Concatenate bytes in order               |

Aggregation queries always result in a single key-value.
With non-aggregation queries, variables and the `...` token
are resolved as actual data elements in the query results.
For aggregation queries, only aggregation variables are
resolved.

```language-fql {.query}
/deltas("group A",<int>)
```

```language-fql {.result}
/deltas("group A",20)=nil
/deltas("group A",-18)=nil
/deltas("group A",3)=nil
```

```language-fql {.query}
/deltas("group A",<sum>)
```

```language-fql {.result}
/deltas("group A",5)=nil
```

The `append` pseudo type is useful when [storing large
blobs][]. The data is usually split into chunks stored in
separate key-values. The respective keys contain the byte
offset of each chunk.

[storing large blobs]: https://apple.github.io/foundationdb/blob.html

```language-fql {.query}
/blob(
  "my file",    % The identifier of the blob.
  <offset:int>, % The byte offset within the blob.
)=<chunk:bytes> % A chunk of the blob.
```

```language-fql {.result}
/blob("my file",0)=10e3_bytes
/blob("my file",10000)=10e3_bytes
/blob("my file",20000)=2.7e3_bytes
```

> Instead of printing the actual byte strings in these
> results, only the byte lengths are printed. See
> [Formatting](#formatting) for more details.

Using `append`, the client obtains the entire blob instead
of having to concatenate the chunks themselves.

```language-fql {.query}
/blob("my file",...)=<blob:append>
```

```language-fql {.result}
/blob("my file",...)=22.7e3_bytes
```

# Implementations

FQL defines the query language but leaves certain details
to the implementation. These include connection configuration,
write permissions, transaction boundaries, variable scope,
and result formatting.

## Connection

An implementation determines how users connect to a
FoundationDB cluster. This may involve selecting from
pre-defined cluster files or specifying a custom path.
An implementation could even simulate an FDB cluster
locally for testing purposes.

## Writes

An implementation may disallow write queries unless a
specific configuration option is enabled. This provides
a safeguard against accidental mutations.

## Transactions

An implementation defines how transaction boundaries are
specified. The Go implementation uses CLI flags to group
queries into transactions.

```language-bash
$ fql \
  -q /users(100)="Alice" \
  -q /users(101)="Bob" \
  --tx \
  -q /users(...)
```

The `--tx` flag represents a transaction boundary. The
first two queries execute within the same transaction.
The third query runs in its own transaction.

## Variables

An implementation defines the scope of named variables.
Variables may be namespaced to a single transaction,
available across multiple transactions, or persist for
an entire session.

## Formatting

An implementation can provide multiple formatting options
for key-values returned by read queries. The default format
prints key-values as their equivalent write queries. This
means copy-pasting the result of a read would write all the
key-values that were read.

Alternative formats may be provided for different use cases:

- Print byte lengths instead of actual bytes to reduce
  output verbosity for large values
- Print placeholders (`<uuid>`, `<vstamp>`) in place of
  actual values when the specific values are not relevant
- Output key-values in a binary format suitable for storage
  on disk or transmission over a network

## Extensions

An implementation may provide custom options and pseudo
tokens beyond those defined by FQL. For example, a predefined
reference `:rand` could generate a random integer for each
query. A custom option `pick:5` could filter results to
return only every fifth key-value.

# Grammar

The complete FQL grammar is specified below.

```language-ebnf {.grammar}
(* Top-level query structure *)
query = [ opts nl ] ( keyval | key | dquery )
dquery = directory [ '=' 'remove' ]

keyval = key '=' value
key = directory tuple
value = 'clear' | data

(* Directories *)
directory = '/' ( '<>' | name | string ) [ directory ]

(* Tuples *)
tuple = '(' [ nl elements [ ',' ] nl ] ')'
elements = data [ ',' nl elements ] | '...'

(* Data elements *)
data = 'nil' | bool | int | num | string | uuid
     | bytes | tuple | vstamp | hole

bool = 'true' | 'false'
int = [ '-' ] digits
num = int '.' digits | ( int | int '.' digits ) 'e' int
string = '"' { char | '\"' } '"'
uuid = hex{8} '-' hex{4} '-' hex{4} '-' hex{4} '-' hex{12}
bytes = '0x' { hex hex }
vstamp = '#' [ hex{20} ] ':' hex{4}

(* Holes: '...' is semantically a hole but defined in tuple to restrict it's usage *)
hole = variable | reference
variable = '<' [ name ':' ] [ type { '|' type } ] '>'
reference = ':' name
type = 'any' | 'tuple' | 'bool' | 'int' | 'num'
     | 'str' | 'uuid' | 'bytes' | 'vstamp' | agg
agg = 'count' | 'sum' | 'avg' | 'min' | 'max' | 'append'

(* Options *)
opts = '[' option { ',' option } ']'
option = name [ ':' argument ]
argument = name | int | string

(* Primitives *)
digits = digit { digit }
digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
hex = digit | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'
    | 'A' | 'B' | 'C' | 'D' | 'E' | 'F'
name = ( letter | '_' ) { letter | digit | '_' | '-' | '.' }
letter = 'a' | ... | 'z' | 'A' | ... | 'Z'
char = ? Any printable ASCII character except '"' ?

(* Whitespace *)
ws = { ' ' | '\t' }
nl = { ' ' | '\t' | '\n' | '\r' }
```

<!-- vim: set tw=60 :-->
