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

# Overview

FQL is specified as a context-free [grammar][]. The
queries look like key-values encoded using the directory
& tuple layers.

[grammar]: https://github.com/janderland/fql/blob/main/syntax.ebnf

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

The query above has a variable `<int>` as its value.
Variables act as placeholders for any of the supported [data
elements](#data-elements).

FQL queries may also perform range reads & filtering by
including a variable in the key's tuple. The query below
will return all key-values which conform to the schema
defined by the query.

```language-fql {.query}
/my/directory(<>,"tuple")=nil
```

```language-fql {.result}
/my/directory("your","tuple")=nil
/my/directory(42,"tuple")=nil
```

All key-values with a certain key prefix may be range read
by ending the key's tuple with `...`.

```language-fql {.query}
/my/directory("my","tuple",...)=<>
```

```language-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

A query's value may be omitted to imply a variable, meaning
the following query is semantically identical to the one
above.

```language-fql {.query}
/my/directory("my","tuple",...)
```

```language-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

Including a variable in the directory tells FQL to perform
the read on all directory paths matching the schema.

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

# Data Elements

An FQL query contains instances of data elements. These
mirror the types of elements found in the [tuple layer][].

[tuple layer]: https://github.com/apple/foundationdb/blob/main/design/tuple.md

<div>

| Type    | Description    | Example                                |
|:--------|:---------------|:---------------------------------------|
| `nil`   | Empty Type     | `nil`                                  |
| `bool`  | Boolean        | `true` `false`                         |
| `int`   | Signed Integer | `-14` `3033`                           |
| `num`   | Floating Point | `33.4` `-3.2e5`                        |
| `str`   | Unicode String | `"happy😁"` `"\"quoted\""`             |
| `uuid`  | UUID           | `5a5ebefd-2193-47e2-8def-f464fc698e31` |
| `bytes` | Byte String    | `0xa2bff2438312aac032`                 |
| `tup`   | Tuple          | `("hello",27.4,nil)`                   |

</div>

The `nil` type may only be instantiated as the element
`nil`. The `int` type may be instantiated as any arbitrarily
large integer. 

```
/int(9223372036854775808)=nil
```

The `num` type may be instantiated as any real number
between `-1.18e4932` and `1.18e4932`, and may use scientific
notation. The type may also be instantiated as the tokens
`-inf`, `inf`, `-nan`, or `nan`. The element is represented
as an 80-bit extended double [floating-point][] and will
snap to the nearest representable number.

[floating-point]: https://en.wikipedia.org/wiki/Extended_precision#x86_extended_precision_format

```language-fql {.query}
/float(-inf,nan)=1.234e4732
```

The `str` type is the only element type allowed in directory
paths. If a directory string only contains alphanumericals,
underscores, dashes, and periods then the quotes may not be
included.

```language-fql {.query}
/quoteless-string_in.dir(true)=false
/"other ch@r@cters must be quoted!"(20)=32.3
```

Quoted strings may contain quotes via backslash escapes.

```language-fql {.query}
/escape("I said \"hello\"")=nil
```

The hexidecimal numbers of the `uuid` and `bytes` types may
be upper, lower, or mixed case.

```language-fql {.query}
/hex(fC2Af671-a248-4AD6-ad57-219cd8a9f734)=0x3b42ADED28b9
```

The `tup` type may contain any of the data elements,
including sub-tuples. Like tuples, a query's value may
contain any of the data elements.

```language-fql {.query}
/sub/tuple("japan",("sub",nil))=0xff
/tuple/value(22.3,-8)=("rain","fog")
```

# Element Encoding

For data elements in the key, the directory and tuple layers
are responsible for data encoding. In the value, the tuple
layer may be used, but FQL supports other encodings known as
"raw values".

```
/tuple_value()={4000}
/raw_value()=4000
```

As a raw value, the `int` type doesn't support an encoding
for arbitrarily large integers. As a value, you'll need to
encode such integers using the tuple layer.

```language-fql {.query}
/int()={9223372036854775808}
```

Below, you can see the default encodings of each type when
used as a raw value.

<div>

| Type    | Encoding                           |
|:--------|:-----------------------------------|
| `nil`   | empty byte array                   |
| `bool`  | single byte, `0x00` means false    |
| `int`   | 64-bit, 1's compliment, big endian |
| `num`   | 64-bit, IEEE 754, big endian       |
| `str`   | UTF-8                              |
| `uuid`  | RFC 4122                           |
| `bytes` | as provided                        |
| `tup`   | tuple layer                        |

</div>

The tuple layer supports a unique encoding for `nil`, but as
a raw value `nil` is equivalent to an empty byte array. This
makes the following two queries equivalent.

```language-fql {.query}
/entry(537856)=nil
/entry(537856)=0x
```

Whether encoded using the tuple layer or as a raw value, the
`int` and `num` types support several different encodings.
A non-default encoding may be specified using the
[options](#options) syntax. Options are specified in
a braced list after the element. If the element's value
cannot be represented by specified encoding then the query
is invalid.

```language-fql {.query}
/numbers(362342[i16])=32.55[f32]
```

By default, [variables](#holes-&-schemas) will decode any
encoding for its types. Options may be applied to
a variable's types to limit which encoding will match the
schema.

```language-fql {.query}
/numbers(<int[i16,big]>)=<num[f32]>
```

The tables below shows which options are supported for the
`int` and `num` types.

<div>

| Int Option | Description     |
|:-----------|:----------------|
| `big`      | Big endian      |
| `lil`      | Little Endian   |
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

| Num Options | Description   |
|:------------|:--------------|
| `big`       | Big endian    |
| `lil`       | Little Endian |
| `f32`       | 32-bit        |
| `f64`       | 64-bit        |
| `f80`       | 80-bit        |

</div>

# Holes & Schemas

A hole is any of the following syntax constructs: variables,
references, and the `...` token. Holes are used to define
a key-value schema by acting as placeholders for one or more
data elements.

A single [data element](#data-elements) may be represented
with a variable. Variables are specified as a list of
element types, separated by `|`, wrapped in angled braces.

```language-fql
<int|str|uuid|bytes>
```

The variable's type list describes which data elements are
allowed at the variable's position. A variable may be empty,
including no element types, meaning it represents all
element types.

```language-fql {.query}
/user(<int>,<str>,<>)=<>
```

```language-fql {.result}
/user(0,"jon",0xffab0c)=nil
/user(20,"roger",22.3)=0xff
/user(21,"",nil)=nil
```

Before the type list, a variable may be given a name. This
name is used to reference the variable in subsequent
queries, allowing for [index indirection](#indirection).

```language-fql {.query}
/index("cars",<varName:int>)
/data(:varName,...)
```

```language-fql {.result}
/user(33,"mazda")=nil
/user(320,"ford")=nil
/user(411,"chevy")=nil
```

Named variables must include at least one type. To allow
named variables to match any element type, use the `any`
type.

```language-fql
/stuff(<thing:any>)
/count(:thing,<int>)
```

```language-fql {.result}
/count("cat",10)
/count(42,1)
/count(0x5fae,3)
```

The `...` token represents any number of data elements of
any type.

```language-fql
/tuples(0x00,...)
```

```language-fql {.result}
/tuples(0x00,"something")=nil
/tuples(0x00,42,43,44)=0xabcf
/tuples(0x00)=nil
```

> ❓ Currently, the `...` token is only allowed as the last
> element of a tuple. This will be revisited in the future.

# Space & Comments

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
line. They can be used to describe a tuple's elements.

```language-fql
% private account balances
/account/private(
  <int>,  % group ID
  <int>,  % account ID
  <str>,  % account name
)=<int>   % balance in USD
```

# Options

Options provide a way to modify the default behavior of data
elements, variable types, and queries. Options are specified
as a comma separated list wrapped in braces.

For instance, to specify that an `int` should be encoded as
a little-endian unsigned 8-bit integer, the following
options would be included after the number.

```language-fql
3548[u8,lil]
```

Similarly, if a variable should only match against
a big-endian 32-bit float then the following option would be
included after the `num` type.

```language-fql
<num[f32,big]>
```

Query options are specified on the line before the query.
For instance, to specify that a range-read query should read
in reverse and only read 5 items, the following options
would be included before the query.

```language-fql {.query}
[reverse,limit:5]
/my/integers(<int>)=nil
```

Notice that the `limit` option includes an argument after
the colon. Some options include a single argument to further
specify the option's behavior.

# Basic Queries

FQL queries may mutate a single key-value, read one or more
key-values, or list directories. Throughout this section,
snippets of Go code are included which approximate how the
queries interact with the FoundationDB API.

## Mutations

Queries lacking [holes](#holes-schemas) perform mutations on
the database by either writing or clearing a key-value.

> ❗ Queries lacking a value altogether imply an empty
> [variable](#holes-schemas) as the value and should not be
> confused with mutation queries.

Mutation queries with a [data element](#data-elements) as
their value perform a write operation.

```language-fql {.query}
/my/dir("hello","world")=42
```

```lang-go {.equiv-go}
db.Transact(func(tr fdb.Transaction) (interface{}, error) {
  dir, err := directory.CreateOrOpen(tr, []string{"my", "dir"}, nil)
  if err != nil {
    return nil, err
  }

  val := make([]byte, 8)
  // Endianness is configurable...
  binary.LittleEndian.PutUint64(val, 42)

  tr.Set(dir.Pack(tuple.Tuple{"hello", "world"}), val)
  return nil, nil
})
```

Mutation queries with the `clear` token as their value
perform a clear operation.

```language-fql {.query}
/my/dir("hello","world")=clear
```

```lang-go {.equiv-go}
db.Transact(func(tr fdb.Transaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"my", "dir"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  tr.Clear(dir.Pack(tuple.Tuple{"hello", "world"}))
  return nil, nil
})
```

## Reads

Queries containing [holes](#holes-schemas) read one or more
key-values. If the holes only appears in the value, then
a single key-value is returned, if one matching the schema
exists.

> ❗ Queries lacking a value altogether imply an empty
> [variable](#holes-schemas) as the value which makes them
> read queries.

```language-fql {.query}
/my/dir(99.8,7dfb10d1-2493-4fb5-928e-889fdc6a7136)=<int|str>
```

```lang-go {.equiv-go}
db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"my", "dir"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  // Read the value's raw bytes...
  val := tr.MustGet(dir.Pack(tuple.Tuple{99.8,
    tuple.UUID{
      0x7d, 0xfb, 0x10, 0xd1,
      0x24, 0x93, 0x4f, 0xb5,
      0x92, 0x8e, 0x88, 0x9f,
      0xdc, 0x6a, 0x71, 0x36}))

  // Try to decode the value as a uint.
  if len(val) == 8 {
      return binary.LittleEndian.Uint64(val), nil
  }

  // If the value isn't a uint, assume it's a string.
  return string(val), nil
})
```

FQL attempts to decode the value as each of the types listed
in the variable, stopping at first success. If the value
cannot be decoded, the key-value does not match the schema.

If the value is specified as an empty variable, then the raw
bytes are returned.

```language-fql {.query}
/some/data(10139)=<>
```

```lang-go {.equiv-go}
db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"some", "data"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  // No value decoding...
  return tr.MustGet(dir.Pack(tuple.Tuple{10139})), nil
})
```

Queries with [variables](#holes-schemas) in their key (and
optionally in their value) result in a range of key-values
being read.

```language-fql {.query}
/people("coders",...)
```

```lang-go {.equiv-go}
db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"people"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  rng, err := fdb.PrefixRange(dir.Pack(tuple.Tuple{"coders"}))
  if err != nil {
    return nil, err
  }

  var results []fdb.KeyValue
  iter := tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
  for iter.Advance() {
    kv := iter.MustGet()

    tup, err := dir.Unpack(kv.Key)
    if err != nil {
      return nil, err
    }

    results = append(results, kv)
  }
  return results, nil
})
```

## Directories

The directory layer may be queried in isolation by using
a lone directory as a query. These queries can only perform
reads. If the directory path contains no variables, the
query will read that single directory.

```language-fql {.query}
/root/<>/items
```

```lang-go {.equiv-go}
 root, err := directory.Open(tr, []string{"root"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  oneDeep, err := root.List(tr, nil)
  if err != nil {
    return nil, err
  }

  var results [][]string
  for _, dir1 := range oneDeep {
    twoDeep, err := root.List(tr, []string{dir1, "items"})
    if err != nil {
      return nil, err
    }

    for _, dir2 := range twoDeep {
      results = append(results, []string{"root", dir1, dir2})
    }
  }
  return results, nil
```

## Filtering

Read queries define a schema to which key-values may or
may-not conform. In the Go snippets above, non-conformant
key-values were being filtered out of the results.

Alternatively, FQL can throw an error when encountering
non-conformant key-values. This may help enforce the
assumption that all key-values within a directory conform to
a certain schema.

TODO: Link to FQL options.

Because filtering is performed on the client side, range
reads may stream a lot of data to the client while the
client filters most of it away. For example, consider the
following query:

```language-fql {.query}
/people(3392,<str|int>,<>)=(<uint>,...)
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
can see a Go implementation of how this filtering would
work.

```lang-go
db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"people"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  rng, err := fdb.PrefixRange(dir.Pack(tuple.Tuple{3392}))
  if err != nil {
    return nil, err
  }

  var results []fdb.KeyValue
  iter := tr.GetRange(rng, fdb.RangeOptions{}).Iterator()
  for iter.Advance() {
    kv := iter.MustGet()

    tup, err := dir.Unpack(kv.Key)
    if err != nil {
      return nil, err
    }

    // Our query specifies a key-tuple
    // with 3 elements...
    if len(tup) != 3 {
      continue
    }

    // The 2nd element must be either a
    // string or an int64...
    switch tup[1].(type) {
    default:
      continue
    case string | int64:
    }

    // The query tells us to assume the value
    // is a packed tuple...
    val, err := tuple.Unpack(kv.Value)
    if err != nil {
      continue
    }

    // The value-tuple must have one or more
    // elements in it...
    if len(val) == 0 {
      continue
    }

    // The first element of the value-tuple must
    // be a uint64...
    if _, isInt := val[0].(uint64); !isInt {
      continue
    }

    results = append(results, kv)
  }
  return results, nil
})
```

# Advanced Queries

Besides basic CRUD operations, FQL is capable of performing
indirection and aggregation queries.

## Indirection

> 🚧 Indirection is still being implemented.

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

## Aggregation

> 🚧 Aggregation is still being implemented.

Aggregation queries read multiple key-values and combine
them into a single output key-value.

FoundationDB performs best when key-values are kept small.
When storing large [blobs][], the blobs are usually split
into 10 kB chunks and stored as values. The respective keys
contain the byte offset of the chunks.

[blobs]: https://apple.github.io/foundationdb/blob.html

```language-fql {.query}
/blob(
  "audio.wav",  % The identifier of the blob.
  <offset:int>, % The byte offset within the blob.
)=<chunk:bytes> % A chunk of the blob.
```

```language-fql {.result}
/blob("audio.wav",0)=10000_bytes
/blob("audio.wav",10000)=10000_bytes
/blob("audio.wav",20000)=2730_bytes
```

> ❓ In the above results, instead of printing the actual
> byte strings, only the byte lengths are printed. This is
> an option provided by the CLI to lower result verbosity.

This gets the job done, but it would be nice if the client
could obtain the entire blob as a single byte string. This
can be done using aggregation queries.

FQL provides a pseudo type named `append` which instructs
the query to append all byte strings found at the variable's
location.

```language-fql {.query}
/blob("audio.wav",...)=<append>
```

```language-fql {.result}
/blob("my file",...)=22730_bytes
```

Aggregation queries always result in a single key-value.
Non-aggregation queries resolve variables & the `...` token
into actual data elements in the query results. Aggregation
queries only resolve aggregation variables.

You can see all the supported aggregation types below.

| Pseudo Type | Accepted Inputs | Description      |
|:------------|:----------------|:-----------------|
| `append`    | `bytes` `str`   | Append arrays    |
| `sum`       | `int` `num`     | Add numbers      |
| `count`     | `any`           | Count key-values |

# Using FQL

The FQL project provides an application for executing
queries and exploring the data, similar to `psql` for
Postgres. This libraries powering this application are
exposed as a Go API, allowing FQL to be used as a Foundation
DB [layer][];

[layer]: https://apple.github.io/foundationdb/layer-concept.html

## Command Line

<div class="language-bash">

### Headless

FQL provides a CLI for performing queries from the command
line. To execute a query in "headless" mode (without
fullscreen), you can use the `-q` flag. The query following
the `-q` flag must be wrapped in single quotes to avoid
mangling by BASH.

```language-bash
ᐅ fql -q '/my/dir("hello","world")'
/my/dir("hello","world")=nil
```

The `-q` flag may be provided multiple times. All queries
are run within a single transaction.

```language-bash
ᐅ fql -q '/my/dir("hello",<var:str>)' -q '/other(22,...)'
/my/dir("hello","world")=nil
/other(22,"1")=0xa8
/other(22,"2")=0xf3
```

### Fullscreen

If the CLI is executed without the `-q` flag, a fullscreen
environment is started up. Single queries may be executed in
their own transactions and the results are displayed in
a scrollable list.

![](img/demo.gif)

Currently, this environment is not very useful, but it lays
the groundwork for a fully-featured FQL frontend. The final
version of this environment will provide autocompletion,
querying of locally cached data, and display customizations.

</div>

## Programmatic

FQL exposes it's AST as an API, allowing Go applications to
use FQL as an FDB layer. The `keyval` package can be used to
construct queries in a partially type-safe manner. While
many invalid queries are caught by the Go type system,
certain queries will only error at runtime.

```language-go
import kv "github.com/janderland/fql/keyval"

var query = kv.KeyValue{
  Key: kv.Key{
    Directory: kv.Directory{
      kv.String("user"),
      kv.String("entry"),
    },
    Tuple: kv.Tuple{
      kv.Int(22573),
      kv.String("Goodwin"),
      kv.String("Samuels"),
    },
  },
  Value: kv.Nil{},
}
```

The `facade` package wraps the FDB client with an
indirection layer, allowing FDB to be mocked. Here we
initialize the default implementation of the facade.
A global root directory is provided at construction time.

```language-go
import (
  "github.com/apple/foundationdb/bindings/go/src/fdb"
  "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
  "github.com/apple/foundationdb/bindings/go/src/tuple"

  "github.com/janderland/fql/engine/facade"
)

func _() {
  fdb.MustAPIVersion(620)
  db := facade.NewTransactor(
    fdb.MustOpenDefault(), directory.Root()))

  db.Transact(func(tr facade.Transaction) (interface{}, error) {
    dir, err := tr.DirOpen([]string{"my", "dir"})
    if err != nil {
      return nil, err
    }
    return nil, tr.Set(dir.Pack(tuple.Tuple{"hi", "world"}, nil)
  })
}
```

The `engine` package executes FQL queries. Each of the five
types of queries has it's own method, making the intended
operation explicit. If a query is used with the wrong
method, an error is returned.

```language-go
import (
  "github.com/apple/foundationdb/bindings/go/src/fdb"
  "github.com/apple/foundationdb/bindings/go/src/fdb/directory"

  "github.com/janderland/fql/engine"
  "github.com/janderland/fql/engine/facade"
  kv "github.com/janderland/fql/keyval"
)

func _() {
  fdb.MustAPIVersion(620)
  eg := engine.New(
    facade.NewTransactor(fdb.MustOpenDefault(), directory.Root()))

  dir := kv.Directory{kv.String("hello"), kv.String("there")}
  key := kv.Key{dir, kv.Tuple{kv.Float(33.3)}}

  // Write: /hello/there{33.3}=10
  query := kv.KeyValue{key, kv.Int(10)}
  if err := eg.Set(query); err != nil {
    panic(err)
  }

  keyExists, err := eg.Transact(func(eg engine.Engine) (interface{}, error) {
    // Write: /hello/there{42}="hello"
    query := kv.KeyValue{
      kv.Key{dir, kv.Tuple{kv.Int(42)}},
      kv.String("hello"),
    }
    if err := eg.Set(query); err != nil {
      return nil, err
    }

    // Read: /hello/there{33.3}=<>
    query = kv.KeyValue{key, kv.Variable{}}
    result, err := eg.ReadSingle(query, engine.SingleOpts{})
    if err != nil {
      return nil, err
    }
    return result != nil, nil
  })
  if err != nil {
    panic(err)
  }
  
  if !keyExists.(bool) {
    panic("keyExists should be true")
  }
}
```

# Roadmap

By summer of 2025, I'd like to have the following items
completed:

- Implement all features described in this document.

- Design and document the syntax for doing the following
  features.

  - Separating queries into multiple transactions.

  - Meta language for aliasing queries or parts of queries.
    This language would provide type-safe templating with
    the goal of reducing repetition in a query file.

Looking beyond summer 2025, I'd like to focus on the TUI
environment:

- Autocompletion and syntax highlighting.

- Query on the results of a previously run query. This would
  allow the user to cache subspaces of data in local memory
  and refine their search with subsequent queries.

- Mechanisms for controlling the output format. These would
  control what is done with the key-values. They could be
  used to print only the first element of the key's tuple or
  to store all the resulting key-values in a flat buffer.
