---
title: FQL
---

```fql {.query}
/user/index/surname("Johnson",<userID:int>)
/user(:userID,...)
```

```fql {.result}
/user(9323,"Timothy","Johnson",37)=nil
/user(24335,"Andrew","Johnson",42)=nil
/user(33423,"Ryan","Johnson",0x0ffa83,42.2)=nil
```

FQL is an [open source] query language and alternative
client API for [FoundationDB]. It's semantics mirror
FoundationDB's [core data model] while improving API
ergonomics. Fundamental patterns like range-reads and
indirection are first class citizens.

[open source]: https://github.com/janderland/fql
[FoundationDB]: https://www.foundationdb.org/
[core data model]: https://apple.github.io/foundationdb/data-modeling.html

<!--toc:start-->
- [Introduction](#introduction)
- [Syntax](#syntax)
  - [Overview](#overview)
  - [Data Elements](#data-elements)
  - [Names](#names)
  - [Directories](#directories)
  - [Schemas](#schemas)
    - [Holes](#holes)
    - [References](#references)
  - [Space & Comments](#space-comments)
  - [Options](#options)
  - [Meta Statements](#meta-statements)
- [Semantics](#semantics)
  - [Data Encoding](#data-encoding)
    - [Keys](#keys)
    - [Values](#values)
    - [Empty](#empty)
    - [Options](#options-1)
  - [Types of Queries](#types-of-queries)
    - [Writes](#writes)
    - [Reads](#reads)
    - [Directories](#directories-1)
    - [Filtering](#filtering)
    - [Options](#options-2)
  - [Advanced Queries](#advanced-queries)
    - [Versionstamps](#versionstamps)
    - [Indirection](#indirection)
    - [Aggregation](#aggregation)
- [Implementations](#implementations)
  - [Connection](#connection)
  - [Permissions](#permissions)
  - [Transactions](#transactions)
  - [Variables & References](#variables-references)
  - [Extensions](#extensions)
  - [Formatting](#formatting)
- [Grammar](#grammar)
<!--toc:end-->

# Introduction 

---

FoundationDB provides the *foundations* of a fully-featured
ACID, distributed, key-value database. It implements
solutions for the hard problems related to distributed data
sharding and replication. Highly concurrent workflows are
enabled via many small, lock-free transactions. Key-values
are stored in sorted order and large batches of adjacent
key-values can be efficiently streamed to clients.

Traditionally, client access is facilitated by a low level
C library with various language bindings. FQL is a [layer]
atop this library, providing a query language and
a higher-level client API. FQL provides a generic way of
describing and querying FoundationDB data, facilitating
schema documentation, client implementation, and debugging.

[layer]: https://apple.github.io/foundationdb/layer-concept.html

This document serves as both a language specification and
a usage guide for FQL. The [Syntax] section describes the
structure of queries while the [Semantics] section describes
their behavior. The [Implementations] section describes the
Go reference implementation and highlights details not
dictated by the specification. The complete [EBNF grammar]
appears at the end. 

> ❗ Not all features described in this document have been
> implemented yet. See the project's [issues] for a roadmap
> of implementation plans.

[issues]: https://github.com/janderland/fql/issues

# Syntax

---
 
Throughout this section, relevant grammar rules are shown
alongside the text. These rules are written in extended
Backus-Naur form as defined in ISO/IEC 14977 with
a modification: concatenation and rule termination are
implicit.

## Overview

FQL is specified as a context-free [grammar]. The queries
resemble key-values encoded using the [directory] and
[tuple] layers.

Directories are used to group sets of key-values. Often,
though not necessarily, the key-values of a particular
directory will follow the same schema. In this sense, they
are analogous to SQL tables.

Tuples provide a way to encode primitive data types into
byte strings while preserving type information and natural
ordering. For instance, after being serialized and sorted,
the tuple `(22,"abc",false)` will appear before the tuple
`(23,"bcd",true)`.

[directory]: https://apple.github.io/foundationdb/developer-guide.html#directories
[tuple]: https://apple.github.io/foundationdb/data-modeling.html#data-modeling-tuples

```ebnf {.grammar}
query = [ opts '\n' ] ( keyval | key | dquery )
dquery = directory [ '=' 'remove' ]
keyval = key '=' value
key = directory tuple
value = 'clear' | data 
```

To the left of the `=` is the key which includes a directory
path and tuple. To the right is the value. For now, the
`opts`{.hljs-variable} prefixing the query can be ignored.
[Options] will be described later in the document.

A query may be a full key-value, just a key, or just
a directory path. The contents of the query implies whether
it's reading or writing data.

```fql {.query}
/my/directory("my","tuple")=4000
```

FQL queries may define a single key-value to be written, as
shown above, or may define a set of key-values to be read,
as shown below.

```fql {.query}
/my/directory("my","tuple")=<int>
```

```fql {.result}
/my/directory("my","tuple")=4000
```

The query above has the variable `<int>` as its value.
Variables act as placeholders for any of the supported [data
elements]. 

FQL queries may also perform [range reads] and filtering by
including one or more variables in the key. The query below
will return all key-values which conform to the schema it
defines.

[range reads]: https://apple.github.io/foundationdb/developer-guide.html#range-reads

```fql {.query}
/my/directory(<>,"tuple")=nil
```

```fql {.result}
/my/directory("your","tuple")=nil
/my/directory(42,"tuple")=nil
```

Unlike the first variable we saw, the variable `<>` in the
query above lacks a type. This means the schema allows any
type of [data element] at the variable's position.

All key-values with a certain key prefix may be range read
by ending the key's tuple with `...`. Due to sorting,
key-values with a common prefix are stored adjacently and
are efficiently streamed to the client.

```fql {.query}
/my/directory("my","tuple",...)=<>
```

```fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

A query's value may be omitted to imply the variable `<>`,
meaning the following query is semantically identical to the
one above.

```fql {.query}
/my/directory("my","tuple",...)
```

```fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

Key-values may be cleared by using the special `clear` token
as the value. If the schema matches multiple keys they will
all be cleared by the query.

```fql {.query}
/my/directory("my",...)=clear 
```

Including a variable in the directory path tells FQL to
perform the read on all directory paths matching the schema.

```fql {.query}
/<>/directory("my","tuple")
```

```fql {.result}
/my/directory("my","tuple")=0x0fa0
/your/directory("my","tuple")=nil
```

The directory path may end with the `...` token to perform
the read on all descendant directories.

```fql {.query}
/your/...(...)
```

```fql {.result}
/your/directory("my","tuple")=nil
/your/keyspace("the","tuple")=547
/your/keyspace/subspace("tuple")="value"
```

The directory layer may be queried by only including
a directory path.

```fql {.query}
/my/<>
```

```fql {.result}
/my/directory
```

Directories are not explicitly created. During a write
query, the directory is created if it doesn't exist.
Directories, along with all their contained key-values, may
be explicitly removed by suffixing the directory path with
`=remove`.

```fql {.query} 
/my/directory=remove
```

## Data Elements

An FQL query contains instances of data elements. These
mirror the types of elements found in the [tuple layer].
This section describes how data elements behave in FQL,
while [element encoding] describes how FQL encodes the
elements before writing them to the DB.

[tuple layer]: https://github.com/apple/foundationdb/blob/main/design/tuple.md

<div>

| Type     | Description      | Examples                               |
|:---------|:-----------------|:---------------------------------------|
| `nil`    | Empty Type       | `nil`                                  |
| `bool`   | Boolean          | `true` `false`                         |
| `int`    | Signed Integer   | `-14` `3033`                           |
| `num`    | Floating Point   | `33.4` `-3.2e5`                        |
| `str`    | Unicode String   | `"happy😁"` `"\"quoted\""`             |
| `bytes`  | Byte String      | `0xa2bff2438312aac032`                 |
| `uuid`   | UUID             | `5a5ebefd-2193-47e2-8def-f464fc698e31` |
| `vstamp` | Version Stamp    | `#:0000` `#0102030405060708090a:0000`  |
| `tup`    | Tuple            | `("hello",27.4,nil)`                   |

</div>

The `nil` type may only be instantiated as the element
`nil`.

```ebnf {.grammar}
bool = 'true' | 'false'
```

The `bool` type may be instantiated as `true` or
`false`.

```ebnf {.grammar}
int = [ '-' ] digits
digits = digit { digit }
digit = '0' | ... | '9'
```

The `int` type may be instantiated as any arbitrarily large
integer.

```ebnf {.grammar}
num = int '.' digits
    | ( int | int '.' digits ) 'e' int 
    | '-inf' | 'inf' | '-nan' | 'nan'
```

The `num` type may be instantiated as any real number which
can be approximated by an [80-bit floating point] value, in
accordance with IEEE 754. Scientific notation may be used.
As expressed in the above specification, the type may be
instantiated as the tokens `-inf`, `inf`, `-nan` or `nan`.

[80-bit floating point]: https://en.wikipedia.org/wiki/Extended_precision#x86_extended_precision_format

```ebnf {.grammar}
string = '"' { char | '\\"' | '\\\\' } '"'
char = ? Any printable UTF-8 character except '"' and '\' ?
```

The `str` type may be instantiated as a unicode string
wrapped in double quotes. Strings may contain double quotes
and backslashes via backslash escapes.

```ebnf {.grammar}
uuid = hex{8} '-' hex{4} '-' hex{4} '-' hex{4} '-' hex{12}
bytes = '0x' { hex hex } 
hex = digit | 'a' | ... | 'f' | 'A' | ... | 'F' 
```

The `uuid` and `bytes` types may be instantiated using
upper, lower, or mixed case hexidecimal numbers. For `uuid`,
the numbers must be grouped in the standard 8, 4, 4, 4, 12
format. For `bytes`, any even number of hexidecimal digits
must be prefixed by `0x`.

```ebnf {.grammar}
vstamp = '#' [ hex{20} ] ':' hex{4}
```

The `vstamp` type represents a FoundationDB [versionstamp]
containing a 10-byte transaction version followed by
a 2-byte user version. These byte strings may be
instantiated using upper, lower, or mixed case hexadecimal
digits. The transaction version may be omitted. In this case
it acts as a placeholder where FoundationDB will write the
actual transaction version upon commit (see
[versionstamps]).

[versionstamp]: https://apple.github.io/foundationdb/data-modeling.html?highlight=versionstamp#versionstamps

```ebnf {.grammar}
tuple = '(' [ nl elements [ ',' ] nl ] ')'
elements = data [ ',' nl elements ] | '...'
```

The `tup` type may contain any of the data elements,
including nested tuples. A trailing comma is allowed after
the last element. The last element may be the `...` token
(see [holes]).

## Names

Names are a syntax construct used throughout FQL. They are
not a [data element] because they are *usually* not
serialized and written to the database. They are used in
many contexts including [directories], [options], and
[variables].

```ebnf {.grammar}
name = ( letter | '_' ) { letter | digit | '_' | '-' | '.' }
```

A name must start with a letter or underscore, followed by
any combination of letters, digits, underscores, dashes, or
periods.

## Directories

Directories provide a way to organize key-values into
hierarchical namespaces. The [directory layer] manages
these namespaces and maps each directory path to a short key
prefix. Key-values in the same directory will have the same
key prefix, and therefore be adjacently stored.

[directory layer]: https://apple.github.io/foundationdb/developer-guide.html#directories

```ebnf {.grammar}
directory = '/' element [ directory ]
element = '<>' | name | string
```

A directory is specified as a sequence of strings, each
prefixed by a forward slash. If the string only contains
characters allowed in a [name], the quotes may be
excluded.

```fql {.query}
/my/directory/path_way
/another/"d!r3ct0ry"/"\"path\""
```

The empty variable `<>` may be used in a directory path as
a placeholder, allowing multiple directories to be queried
at once.

```fql {.query}
/app/<>/index
```

```fql {.result}
/app/users/index
/app/roles/index
/app/actions/index
```

## Schemas

### Holes

Holes are used to define a key-value schema by acting as
placeholders for one or more data elements. There are two
kinds of holes: variables and the `...` token.

```ebnf {.grammar}
variable = '<' [ name ':' ] [ type { '|' type } ] '>'
type = 'any' | 'tuple' | 'bool' | 'int' | 'num'
     | 'str' | 'uuid' | 'bytes' | 'vstamp'
```

Variables are used to represent a single [data element].
They may optionally include a unique [name] followed by
their type. The variable below is named "myVar" and acts as
a placeholder for any integer value.

```fql
<myVar:int>
```

A variable may act as a placeholder for multiple types of
elements with the types separated by `|`{.hljs-variable}. It
may also have no type meaning it represents any type of
element.

```fql {.query}
/tree/node(<int>,<int|nil>,<int|nil>)=<>
```

```fql {.result}
/tree/node(5,12,14)=nil
/tree/node(12,nil,nil)="payload"
/tree/node(14,nil,15)=0xa3127b
/tree/node(15,nil,nil)=(42,96,nil)
```

The `...` token represents any number of data elements of
any type. It is only allowed as the last element of a tuple.

```fql {.query}
/app/queue("topic",...)
```

```fql {.result}
/app/queue("topic",54,"process: 12643")
/app/queue("topic",55,"process: 12644")
/app/queue("topic",56,"process: 12648")
/app/queue("topic",57,"process: 12649")
/app/queue("topic",58,"done")
```

### References

```ebnf {.grammar}
reference = ':' name
```

References can use a variable's name to pass previously read
values into a subsequent query, allowing for [index
indirection]. The reference is specified as the variable's
name prefixed with a `:`{.hljs-variable}.

```fql {.query}
/user/index/surname("Johnson",<userID:int>)
/user(:userID,...)
```

```fql {.result}
/user(9323,"Timothy","Johnson",37,"United States")=nil
/user(24335,"Andrew","Johnson",42,"United States")=nil
/user(33423,"Ryan","Johnson",32,"England")=nil
```

Named variables must include at least one type. To allow
named variables to match all element type, use the `any`
type.

```fql {.query}
/store/hash(<bytes>,<thing:any>)
```

```fql {.result}
/store/hash(0x6dc88b,"somewhere we have")=nil
/store/hash(0x8b593b,523.8e90)=nil
/store/hash(0x9ccf9d,"I have yet to find")=nil
/store/hash(0xcd53e8,ca03676e-1c59-4dd4-a7ea-36c90714c2b7)=nil
/store/hash(0xda3924,0x96f70a30)=nil
```

## Space & Comments

Whitespace and newlines are allowed within a tuple, between
its elements.

```fql {.query}
/account/private(
  <int>,
  <int>,
  <str>,
)=<int>
```

Comments start with a `%` and continue until the end of the
line. They can be used to document a tuple's elements.

```fql
% private account balances
/account/private(
  <int>,  % group ID
  <int>,  % account ID
  <str>,  % account name
)=<int>   % balance in USD
```

## Options

Options modify the semantics of [data elements],
[variables], and [queries]. They can instruct FQL to use
alternative encodings, limit a query's result count, or
change other behaviors. 

```ebnf {.grammar}
options = '[' option { ',' option } ']'
option = name [ ':' argument ]
argument = name | int | string
```

Options are specified as a comma separated list wrapped in
brackets. For instance, to specify that an `int` should be
encoded as a little-endian unsigned 8-bit integer, the
following options would be included after the element.

```fql
3548[u8]
```

If a variable should only match against big-endian 32-bit
floats then the following options would be included after
the `num` type.

```fql
<num[f32,be]>
```

Query options are specified on the line before the query. To
specify that a range-read query should read in reverse and
only read 5 items, the following options would be included
before the query.

```fql {.query}
[reverse,limit:5]
/my/integers(<int>)=nil
```

Notice that the `limit` option includes a number after the
colon. Some options require a single argument to further
specify the option's behavior. The argument may be an
integer, a [name], or a string.

Details about the various options will be included in the
sections explaining the semantics which they modify.

## Meta Statements

TODO: @commit

# Semantics 

---

Throughout this section, snippets of Python code are
included showcasing simplified implementations of FQL
features using the FoundationDB API. These snippets don't
include optimizations found in the actual implementation
like concurrency, batching, or caching.

## Data Encoding

FoundationDB stores keys and values as simple byte strings
leaving the client responsible for encoding the data. FQL
determines how to encode [data elements] based on their data
type, position within the query, and associated [options].

### Keys

Keys are *always* encoded using the [directory] and [tuple]
layers. All keys must include a directory prefix. Write
queries create directories if they do not exist.

```fql {.query}
/app/users(57223,"Peter","Carson",56)=nil
```

```python {.equiv-py}
@fdb.transactional
def write_user(tr):
    # Open directory; create if doesn't exist
    dir = fdb.directory.create_or_open(tr, ('app', 'users'))

    # Pack the tuple and prepend the directory prefix
    key = dir.pack((57223, "Peter", "Carson", 56))

    # Encode the value
    val = # ...

    # Write the KV
    tr[key] = val
```

If a query reads from a directory which doesn't exist,
nothing is returned. The tuple layer encodes metadata about
element types, allowing FQL to decode keys without a schema.

```fql {.query}
/app/...(...)
```

```python {.equiv-py}
@fdb.transactional
def read_all(tr):
    # Open directory; exit if it doesn't exist
    dir = fdb.directory.open(tr, ('app',))
    if dir is None:
        return []

    # Recursively read all directories
    return do_read_users(tr, dir)


def do_read_all(tr, dir):
    # Grab all the key-values
    results = []
    for key, val in tr[dir.range()]:
        # Get the full path of the directory
        path = dir.get_path()

        # Remove the directory prefix and unpack the tuple
        tup = dir.unpack(key)

        # Unpack the value
        val = # ...

        # Collect the key-values
        results.append((path, tup, val))

    # Recurse into child directories
    for child_name in dir.list(tr):
        child_dir = dir.open(tr, (child_name,))
        results += do_read_all(tr, child_dir)

    return results
```

### Values

Value [data elements] are encoded as the lone member of
a tuple. This approach preserves type information for
flexible decoding.

```fql {.query}
/people/age("jon","smith")=42
```

```python {.equiv-py}
@fdb.transactional
def write_age(tr):
    # Encoding the key
    key = # ...

    # Pack the value as a tuple
    val = fdb.tuple.pack((42,))

    # Write the key-value
    tr[key] = val
```

```fql {.query}
/people/age("jon","smith")=<>
```

```python {.equiv-py}
@fdb.transactional
def read_age(tr):
    # Encode the key
    key = # ...

    # Read the value's bytes
    val_bytes = tr[key]

    try:
        # Assume the value is a tuple
        val_tup = fdb.tuple.unpack(val_bytes)

        if len(val_tup) == 1:
            # Unwrap single elements
            return val_tup[0]
        else:
            # Return as a tuple
            return val_tup
    except:
        # Fallback to raw bytes
        return val_bytes
```

Not all data elements are encoded the same way. As a value,
tuples are encoded using the tuple layer, but they are not
wrapped in a tuple like the other data elements. Byte
strings are written as-is.

This means that `42` and `(42)` have the same value
encoding. The way the value is returned depends on how it's
queried.

```fql {.query}
% write the key-value once
/app/location("east bay")=87234

% read without a tuple
/app/location("east bay")=<>

% read with a tuple
/app/location("east bay")=(<>)
```

```fql {.result}
/app/location("east bay")=87234
/app/location("east bay")=(87234)
```

### Empty

Within a tuple, `nil`, empty bytes `0x`, and empty nested tuples `()`
are encoded with their types preserved by the [tuple layer].
As a value, all three are encoded as an empty byte string.
A typeless variable will decode an empty byte string as
`nil`.

```fql {.query}
/globals/selection("object")=0x
/globals/selection("item")=nil
/globals/selection("text")=()

/globals/selection(...)=<>
```

```fql {.result}
/globals/selection("object")=nil
/globals/selection("item")=nil
/globals/selection("text")=nil
```

Likewise, the tuple of a key is encoded as an empty byte
string when it contains no elements, allowing queries to use
a key that is simply the directory prefix.

```fql {.query}    
/globals/next-id()=37534
```

```python {.equiv-py}
@fdb.transactional
def set_next_id(tr):
    # Open directory; create if doesn't exist
    dir = fdb.directory.open(tr, ('globals', 'next-id'))

    # Use directory prefix as the key
    key = dir.key()

    # Encode the value
    val = # ...

    # Write key-value
    tr[key] = val
```

### Options

Options allow for encoding values in different ways than
outlined above. The table below shows [options] which change
how the `int` and `num` types are encoded.

> ❗ Options only change how data elements are encoded as
> values. In keys, the tuple layer encodes data elements in
> the smallest representation available.

<div>

| Option      | Argument | Description                            |
|:------------|:---------|:---------------------------------------|
| `width`     | `int`    | Bit width: `8`, `16`, `32`, `64`, `80` |
| `bigendian` | none     | Use big endian encoding                |
| `unsigned`  | none     | Use unsigned encoding                  |

</div>

`int` may use the widths `8`, `16`, `32`, and `64`, while `num` may
use `32`, `64`, and `80`. When the width option is present,
values use little endian encoding, as long as the
`bigendian` option isn't also present.

```fql {.query}
/globals/next-id()=37534[width:64,bigendian]
```

```python {.equiv-py}
@fdb.transactional
def set_next_id(tr):
    # Encode the key
    key = # ...

    # Encode the value as a big-endian, 64-bit, signed int
    val = struct.pack('>q', 37534)

    # Write the key-value
    tr[key] = val
``` 

When writing values with non-default encoding, type metadata
will be lost. Read queries will need the appropriate options
specified. Otherwise, the value will not match the schema.

```fql {.query}
% write
/globals/next-id()=37534[i64,be]

% read
/globals/next-id()=<int[i64,be]>
```

```fql {.result}
/globals/next-id()=37534[i64,be]
```

FQL provides aliases for the `int` and `num` options to
decrease their verbosity. For instance,
`[width:64,bigendian]` can be written as `[i64,be]`. The
tables below list the available aliases for for `int` and
`num` options.

<div>

| Int Alias | Actual Options      |
|:----------|:--------------------|
| `be`      | `bigendian`         |
| `i8`      | `width:8`           |
| `i16`     | `width:16`          |
| `i32`     | `width:32`          |
| `i64`     | `width:64`          |
| `u8`      | `unsigned,width:8`  |
| `u16`     | `unsigned,width:16` |
| `u32`     | `unsigned,width:32` |
| `u64`     | `unsigned,width:64` |

</div>

<div>

| Num Alias | Actual Options |
|:----------|:---------------|
| `be`      | `bigendian`    |
| `f32`     | `width:32`     |
| `f64`     | `width:64`     |
| `f80`     | `width:80`     |

</div>

The `str`, `uuid`, and `vstamp` types include the option
`raw` which causes their bytes to be written as-is without
being wrapped in a tuple. 

```fql {.query}
% write raw UUID
/tag_code("food")=77542869-5708-4af9-821e-d65354fb1a12[raw]

% read as bytes
/tag_code("food")=<bytes>
```

```fql {.result}
/tag_code("food")=0x7754286957084af9821ed65354fb1a12
```

## Types of Queries

FQL queries may write a single key-value, read/clear one or
more key-values, or list/remove directories. As stated
earlier, all queries resemble key-values, and the tokens
within said key-values imply which of the above operations
is executed.

### Writes

Queries lacking [holes] perform writes on
the database. You can think of these queries as declaring
the existence of a particular key-value. If the key's
directory does not exist, it is created during a write
operation.

> ❗ Queries lacking a value altogether imply an empty
> [variable] as the value and should not
> be confused with write queries.

### Reads

Queries containing [holes] read one or more
key-values. If the holes only appear in the value, then
a single key-value is returned, if one matching the schema
exists. Most query results can be fed back into FQL as write
queries. The exception to this rule are [aggregate
queries](#aggregation) and results created by non-default
[formatting].

FQL attempts to decode the value as each of the types listed
in the variable, stopping at first success. If the value
cannot be decoded, the key-value does not match the schema.

Queries with [variables] in their key (and
optionally in their value) result in a range of key-values
being read.

Whether reading single or many, when a key-value is
encountered which doesn't match the query's schema it is
filtered out of the results. Including the `strict` [query
option] causes the query to fail when encountering
a nonconformant key-value.

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

As stated above, read queries define a schema to which
key-values may or may-not conform. Because filtering is
performed on the client side, range reads may stream a lot
of data to the client while filtering most of it away. For
example, consider the following query:

```fql {.query}
/people(3392,<str|int>,<>)=(<int>,...)
```

In the key, the location of the first
[hole] determines the range read prefix
used by FQL. For this particular query, the prefix would be
as follows:

```fql {.query}
/people(3392)
```

FoundationDB will stream all key-values with this prefix to
the client. As they are received, the client will filter out
key-values which don't match the query's schema. This may be
most of the data. Ideally, filter queries are only used on
small amounts of data to limit wasted bandwidth.

Below you can see a Python implementation of how this
filtering would work.

```python
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

### Options

As hinted at above, queries have several options which
modify their default behavior.

<div>

| Query Option | Argument | Description                              |
|:-------------|:---------|:-----------------------------------------|
| `reverse`    | none     | Range read in reverse order              |
| `limit`      | `int`    | Maximum number of results                |
| `mode`       | name     | Range read mode: `want_all`, `iterator`, `exact`, `small`, `medium`, `large`, `serial` |
| `snapshot`   | none     | Use snapshot read                        |
| `strict`     | none     | Error when a read key-values doesn't conform to the schema |

</div>

Range-read queries support all the options listed above.
Single-read queries support `snapshot` and `strict`. Clear
queries support `strict`. With the `strict` option, the
clear operation is a no-op if FQL encounters a key in the
given directory which doesn't match the schema.

## Advanced Queries

### Versionstamps

As stated in the [data elements] section,
a `vstamp` is composed of two components: the transaction
version prefixed by `#`{.hljs-title} and the user version
prefixed by `:`{.hljs-title}.

A `vstamp` lacking a transaction version is called an
"incomplete" `vstamp`. In a write query, an incomplete
`vstamp` has unique behavior. Upon commit, the transaction's
10-byte version is written to the first 10-bytes of the
`vstamp`. 

```fql {.query}
@write
/app/queue(#:ff00)="jason"
/app/heartbeat("jason")=#:00cd

@commit

@read
/app/queue(<index:vstamp>)
/app/heartbeat(...)=<heartbeat:vstamp>
```

```fql {.result}
/app/queue(#8e9ddaa52e44733526e2:ff00)="jason"
/app/heartbeat("jason")=#8e9ddaa52e44733526e3:00cd
```

The example above showcases several details about writing an
incomplete `vstamp`:

- The transaction version component of the `vstamp` is
  written at commit time, so you must start a new
  transaction before reading it. If you attempt to read an
  empty `vstamp` before the transaction is committed, the
  query will fail.

- The user version component of the `vstamp` is not
  overwritten. Only the transaction version is.

- The final two bytes of the transaction version component
  (right before the user version) are incremented within
  a transaction. In this particular example, the `/queue`
  key's transaction version ends with `26e2` while the
  `/heartbeat` transaction version ends with `26e3`. This
  ensures that multiple versionstamps written by the same
  transaction are unique.

`vstamp` elements are monotonically increasing and unique
for the lifetime of a particular database. They may be used
as unique identifiers or non-contiguous indexes.

### Indirection

Indirection queries are similar to SQL joins. They associate
different groups of key-values via some shared data element.

In FoundationDB, indexes are implemented using indirection.
Suppose we have a large list of people, one key-value for
each person.

```fql {.query}
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

```fql {.query}
/people/last_name(
  <str>, % Last Name
  <int>, % ID
)=nil
```

If we query the index, we can get the IDs of the records
containing the last name "Johnson".

```fql {.query}
/people/last_name("Johnson",<int>)
```

```fql {.result}
/people/last_name("Johnson",23)=nil
/people/last_name("Johnson",348)=nil
/people/last_name("Johnson",2003)=nil
```

FQL can forward the observed values of named variables from
one query to the next. We can use this to obtain our desired
subset from the "people" directory.

```fql {.query}
/people/last_name("Johnson",<id:int>)
/people(:id,...)
```

```fql {.result}
/people(23,"Lenny","Johnson",22,"Mechanic")=nil
/people(348,"Roger","Johnson",54,"Engineer")=nil
/people(2003,"Larry","Johnson",8,"N/A")=nil
```

Notice that the results of the first query are not returned.
Instead, they are used to build a collection of single-KV
read queries whose results are the ones returned.

### Aggregation

Aggregation queries combine multiple key-values into
a single key-value. FQL provides pseudo data types for
performing aggregation, similar to SQL's [aggregate
functions].

[aggregate functions]: https://en.wikipedia.org/wiki/Aggregate_function

Suppose we are storing value deltas. If we range-read the
keyspace we end up with a list of integer values.

```fql {.query}
/deltas("group A",<int>)
```

```fql {.result}
/deltas("group A",20)=nil
/deltas("group A",-18)=nil
/deltas("group A",3)=nil
```

Instead, we can use the pseudo type `sum` in our variable to
automatically sum up the deltas into the actual value.

```fql {.query}
/deltas("group A",<sum>)
```

```fql {.result}
/deltas("group A",5)=nil
```

Aggregation queries are also useful when [reading large
blobs][]. The data is usually split into chunks stored in
separate key-values. The respective keys contain the byte
offset of each chunk.

[reading large blobs]: https://apple.github.io/foundationdb/blob.html

```fql {.query}
/blob(
  "my_file.bin",    % The identifier of the blob.
  <offset:int>, % The byte offset within the blob.
)=<chunk:bytes> % A chunk of the blob.
```

```fql {.result}
/blob("my_file.bin",0)=10kb
/blob("my_file.bin",10000)=10kb
/blob("my_file.bin",20000)=2.7kb
```

> ❗ Instead of printing the actual byte strings in these
> results, only the byte lengths are printed. This is
> a possible feature of an FQL implementation. See
> [Formatting] for more details.

Using `append`, the client obtains the entire blob instead
of having to concatenate the chunks themselves.

```fql {.query}
/blob("my_file.bin",...)=<blob:append>
```

```fql {.result}
/blob("my_file.bin",...)=22.7kb
```

With non-aggregation queries, [holes] are
resolved to actual data elements in the results. For
aggregation queries, only aggregation variables are
resolved, leaving the `...` token in the resulting
key-value.

The table below lists the available aggregation types.

<div>

| Aggregate | I/O                            | Description                     |
|:----------|:-------------------------------|:--------------------------------|
| `count`   | `any` ➜ `int`                  | Count the number of results     |
| `sum`     | `int`,`num` ➜ `int`,`num`      | Sum numeric values              |
| `min`     | `int`,`num` ➜ `int`,`num`      | Minimum numeric value           |
| `max`     | `int`,`num` ➜ `int`,`num`      | Maximum numeric value           |
| `avg`     | `int`,`num` ➜ `num`            | Average numeric values          |
| `append`  | `bytes`,`str` ➜ `bytes`,`str`  | Concatenate bytes/strings       |

</div>

`sum`, `min`, and `max` output `int` if all inputs are
`int`. Otherwise, they output `num`. Similarly, `append`
outputs `str` if all inputs are `str`. Otherwise, it outputs
`bytes`.

`append` may be given the [option] `sep` which
defines a `str` or `bytes` separator placed between each of
the appended values.

```fql {.query}
% Append the lines of text for a blog post.
/blog/post(
  253245,      % post ID
  <offset:int> % line offset
)=<body:append[sep:"\n"]>
```

# Implementations

FQL defines the query language but leaves many details to
the implementation. This sections outlines some of those
details and how an implementation may choose to provide
them.

TODO: talk about FQL as a client API.

## Connection

An implementation determines how users connect to a
FoundationDB cluster. This may involve selecting from
predefined cluster files or specifying a custom path.
An implementation could even simulate an FDB cluster
locally for testing purposes.

## Permissions

An implementation may disallow write queries unless a
specific configuration option is enabled. This provides
a safeguard against accidental mutations. Implements could
also limit access to certain directories or any other
behavior for any reason.

## Transactions

An implementation defines how transaction boundaries are
specified. The Go implementation uses CLI flags to group
queries into transactions.

```bash
$ fql \
  -q /users(100)="Alice" \
  -q /users(101)="Bob" \
  --tx \
  -q /users(...)
```

The `--tx` flag represents a transaction boundary. The
first two queries execute within the same transaction.
The third query runs in its own transaction.

## Variables & References

An implementation defines the scope of named variables.
Variables may be namespaced to a single transaction,
available across multiple transactions, or persist for
an entire session.

Named variables could also be used to output specific values
to other parts of the application. For instance, variables
with the name `stdout` may write their values to the STDOUT
stream of the process.

```fql {.query}
/mq("topic",<stdout:str>)
```

```bash {.result}
topicA
topicB
topicC
```

Similarly, references could be used to inject values into
a query from another part of the process.

```fql {.query}
% Write the string contents of STDIN into the DB.
/mq("msg","topicB",:stdin)
```

## Extensions

An implementation may provide custom options and types
beyond those defined by FQL. For example, the pseudo type
`json` could act as a restricted form of `str` which only
matches valid JSON. A custom option `every:5` could filter
results to return only every fifth key-value.

## Formatting

An implementation can provide multiple formatting options
for key-values returned by read queries. The default format
prints key-values as their equivalent write queries.
Alternative formats may be provided for different use cases:

- Print byte lengths instead of actual bytes to reduce
  output verbosity for large values.
- Print placeholders (`<uuid>`, `<vstamp>`) in place of
  actual values when the details are not relevant.
- Output key-values in a binary format suitable for storage
  on disk or transmission over a network.

# Grammar

The complete FQL grammar is specified below.

```ebnf {.grammar}
(* Top-level query structure *)
query = [ opts '\n' ] ( keyval | key | dquery )
dquery = directory [ '=' 'remove' ]

keyval = key '=' value
key = directory tuple
value = 'clear' | data

(* Directories *)
directory = '/' ( '<>' | name | string ) [ directory ]

(* Tuples *)
tuple = '(' [ nl elements [ ',' ] nl ] ')'
elements = '...' | data [ ',' nl elements ]

(* Data elements *)
data = 'nil' | bool | int | num | string | uuid
     | bytes | tuple | vstamp | variable | reference

bool = 'true' | 'false'
int = [ '-' ] digits
num = int '.' digits | ( int | int '.' digits ) 'e' int
string = '"' { char | '\"' | '\\' } '"'
uuid = hex{8} '-' hex{4} '-' hex{4} '-' hex{4} '-' hex{12}
bytes = '0x' { hex{2} }
vstamp = '#' [ hex{20} ] ':' hex{4}

(* Variables and References *)
variable = '<' [ name ':' ] [ type { '|' type } ] '>'
reference = ':' name [ '!' type ]
type = 'any' | 'tuple' | 'bool' | 'int' | 'num'
     | 'str' | 'uuid' | 'bytes' | 'vstamp' | agg
agg = 'count' | 'sum' | 'avg' | 'min' | 'max' | 'append'

(* Options *)
opts = '[' option { ',' option } ']'
option = name [ ':' argument ]
argument = name | int | string

(* Primitives *)
digits = digit { digit }
digit = '0' | '1' | '2' | '3' | '4'
      | '5' | '6' | '7' | '8' | '9'
hex = digit
    | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'
    | 'A' | 'B' | 'C' | 'D' | 'E' | 'F'
name = ( letter | '_' ) { letter | digit | '_' | '-' | '.' }
letter = 'a' | ... | 'z' | 'A' | ... | 'Z'
char = ? Any printable UTF-8 character except '"' and '\' ?

(* Whitespace *)
ws = { ' ' | '\t' }
nl = { ' ' | '\t' | '\n' | '\r' }
```

<!-- vim: set tw=60 :-->
