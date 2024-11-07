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
  [Foundation DB](https://www.foundationdb.org/). It's query
  semantics mirror Foundation DB's [core data
  model](https://apple.github.io/foundationdb/data-modeling.html).
  Fundamental patterns like range-reads and indirection are first
  class citizens.
...

# Overview

FQL is specified as a [context-free
grammar](https://github.com/janderland/fql/blob/main/syntax.ebnf).
The queries look like key-values encoded using the directory
& tuple layers.

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
elements](#data-elements). In this case, the variable also
tells FQL how to decode the value's bytes. This query will
return a single key-value from the database, if such a key
exists.

FQL queries can also perform range reads & filtering by
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

All key-values with a certain key prefix can be range read
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

Key-values can be cleared by using the special `clear` token
as the value.

```language-fql {.query}
/my/directory("my","tuple")=clear
```

The directory layer can be queried by only including
a directory path.

```language-fql {.query}
/my/<>
```

```language-fql {.result}
/my/directory
```

# Data Elements

An FQL query contains instances of data elements. These
mirror the types of elements found in the [tuple
layer](https://github.com/apple/foundationdb/blob/main/design/tuple.md).
Descriptions of these elements can be seen below.

<div>

| Type    | Description    | Example                                |
|:--------|:---------------|:---------------------------------------|
| `nil`   | Empty Type     | `nil`                                  |
| `bool`  | Boolean        | `true`                                 |
| `int`   | Signed Integer | `-14`                                  |
| `num`   | Floating Point | `33.4`                                 |
| `str`   | Unicode String | `"string"`                             |
| `uuid`  | UUID           | `5a5ebefd-2193-47e2-8def-f464fc698e31` |
| `bytes` | Byte String    | `0xa2bff2438312aac032`                 |
| `tup`   | Tuple          | `("hello",27.4,nil)`                   |

</div>

The `nil` type allows for a single value which is also
`nil`. The tuple layer supports a unique encoding for `nil`.
As a value, `nil` is equivalent to an empty byte array, so
the following queries are semantically equivalent:

```language-fql {.query}
/entry(537856)=nil
/entry(537856)=0x
```

The `int` type allows for arbitrarily large integers. The
`num` type allows for 64-bit floating-point numbers. For
now, the `num` type does not support arbitrary precision
because the tuple layer [advises
against](https://github.com/apple/foundationdb/blob/main/design/tuple.md#arbitrary-precision-decimal)
using it's abitrarily-sized decimal encoding. Support for
arbitrarily-sized floating-point numbers will
be revisited in the future.

The `str` type supports unicode strings, including unicode
escape sequences. FQL has not chosen a unicode escape
syntax, but it will be similar to what is found in other
programming languages.

Strings are the only data element allowed in directories. If
a directory string only contains alphanumericals,
underscores, dashes, and periods then the quotes don't need
to be included.

```language-fql {.query}
/quoteless-string_in.dir(true)=false
/"other ch@r@cters must be quoted!"(20)=32.3
```

Quoted strings may contain quotes via backslash escapes.

```language-fql {.query}
/my/dir("I said \"hello\"")=nil
```

The 'tup' type may contain any of the data elements,
including sub-tuples. Like tuples, a query's value may
contain any of the data elements.

```language-fql {.query}
/region/north_america(22.3,-8)=("rain","fog")
/region/east_asia("japan",("sub",nil))=0xff
```

# Value Encoding

The directory and tuple layers are responsible for encoding
the data elements in the key. As for the value, FDB doesn't
provide a standard encoding.

FQL provides default value encoding for each of the data
elements, as show below. The upcoming "options" syntax will
allow queries to specify alternative encodings for each data
element.

<div>

| Type    | Encoding                        |
|:--------|:--------------------------------|
| `nil`   | empty value                     |
| `bool`  | single byte, `0x00` means false |
| `int`   | 64-bit, 1's compliment          |
| `num`   | IEEE 754                        |
| `str`   | Unicode                         |
| `uuid`  | RFC 4122                        |
| `bytes` | as provided                     |
| `tup`   | tuple layer                     |

</div>

# Variables & Schemas

Variables allow FQL to describe key-value schemas. Any [data
element](#data-elements) may be represented with a variable.
Variables are specified as a list of element types,
separated by `|`, wrapped in angled braces.

```language-fql
<uint|str|uuid|bytes>
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
/user(21,"",nil)="nothing"
```

Before the type list, a variable can be given a name. This
name is used to reference the variable in subsequent
queries, allowing for [index
indirection](#index-indirection).

```language-fql {.query}
/index("cars",<varName:int>)
/data(:varName,...)
```

```language-fql {.result}
/user(33,"mazda")=nil
/user(320,"ford")=nil
/user(411,"chevy")=nil
```

# Space & Comments

Whitespace and newlines are allowed within a tuple, between
its elements.

```language-fql {.query}
/account/private(
  <uint>,
  <uint>,
  <str>,
)=<int>
```

Comments start with a `%` and continue until the end of the
line. They can be used to describe a tuple's elements.

```language-fql
% private account balances
/account/private(
  <uint>, % user ID
  <uint>, % group ID
  <str>,  % account name
)=<int>   % balance in USD
```

# Basic Queries

FQL queries can write/clear a single key-value, read one or
more key-values, or list directories. Throughout this
section, snippets of Go code are included to show how the
queries interact with the FDB API.

## Mutations

Queries lacking [variables](#variables) and the `...` token
perform mutations on the database by either writing or
clearing a key-value.

> Queries lacking a value altogether imply an empty
> [variable](#variables) as the value and should not be
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

Queries containing a [variable](#variables) or the `...`
token read one or more key-values. The query defines
a schema which the returned key-values must conform to.

If the variable or `...` token only appears in the query's
value, then it returns a single key-value, if one matching
the schema exists.

> Queries lacking a value altogether imply an empty
> [variable](#variables) as the value, and are therefore
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

Queries with [variables](#variables) or the `...` token in
their key (and optionally in their value) result in a range
of key-values being read.

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

Foundation DB will stream all key-values with this prefix to
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

Besides basic
[CRUD](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete)
operations, FQL is capable of performing indirection and
aggregation queries.

## Indirection

Indirection queries are similar to SQL joins. They associate
different groups of key-values via some shared data element.

In Foundation DB, indexes are implemented by having one
key-value (the index) point at another key-value. This is
also called "indirection".

> Indirection is not yet included in the grammar, nor is it
> implemented. The design of this feature is somewhat
> finalized.

Suppose we have a large list of people, one key-value for
each person.

```language-fql {.query}
/people(<id:uint>,<firstName:str>,<lastName:str>,<age:int>)=nil
```

If we wanted to read all records with the last name of
"Johnson", we'd have to perform a linear search across the
entire "people" directory. To make this kind of search more
efficient, we can store an index of last names in a separate
directory.

```language-fql {.query}
/index/last_name(<lastName:str>,<id:uint>)=nil
```

FQL can forward the observed values of named variables from
one query to the next, allowing us to efficiently query for
all people with the last name of "Johnson".

```language-fql {.query}
/index/last_name("Johnson",<id:uint>)
/people(:id,...)
```
```language-fql {.result}
/people(23,"Lenny","Johnson",22,"Mechanic")=nil
/people(348,"Roger","Johnson",54,"Engineer")=nil
/people(2003,"Larry","Johnson",8,"N/A")=nil
```

The first query returned 3 key-values containing the IDs of
23, 348, & 2003 which were then fed into the second query
resulting in 3 individual [single reads](#single-reads).

```language-fql {.query}
/index/last_name("Johnson",<id:uint>)
```
```language-fql {.result}
/index/last_name("Johnson",23)=nil
/index/last_name("Johnson",348)=nil
/index/last_name("Johnson",2003)=nil
```

## Aggregation

> The design of aggregation queries is not complete. This
> section describes the general idea. Exact syntax may
> change. This feature is not currently included in the
> grammar nor has it been implemented.

Foundation DB performs best when key-values are kept small.
When [storing large
blobs](https://apple.github.io/foundationdb/blob.html), the
data is usually split into 10 kB chunks stored in the value.
The respective key contain the byte offset of the chunk.

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
> results, only the byte lengths are printed. This is an
> option provided by the CLI to lower result verbosity.

This gets the job done, but it would be nice if the client
could obtain the entire blob instead of having to append the
chunks themselves. This can be done using aggregation
queries.

FQL provides a pseudo data type named `agg` which performs
the aggregation.

```language-fql {.query}
/blob("my file",...)=<blob:agg>
```

```language-fql {.result}
/blob("my file",...)=22.7e3_bytes
```

Aggregation queries always result in a single key-value.
With non-aggregation queries, variables & the `...` token
are resolved as actual data elements in the query results.
For aggregation queries, only aggregation variables are
resolved.

A similar pseudo data type for summing integers could be
provided as well.

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
/deltas("group A",5)=<>
```

# Using FQL

The FQL project provides an application for executing
queries and exploring the data, similar to `psql` for
Postgres. This libraries powering this application are
exposed as a Go API, allowing FQL to be used as a Foundation
DB
[layer](https://apple.github.io/foundationdb/layer-concept.html).

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

- Indirection & aggregation queries implemented as described
  in this document.

- Design and document the syntax for doing the following
  features.

  - Separating queries into multiple transactions.

  - Set options at both the query & transaction level.
    Options control things like range-read direction
    & limits, endianness of values, and whether write
    queries are allowed.

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
