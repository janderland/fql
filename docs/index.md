---
title: FQL

# We include this intro via the 'include-before'
# metadata field so it's placed before the TOC.
include-before: |
  ```lang-fql {.query}
  /user/index/surname("Johnson",<userID:int>)
  /user(:userID,...)
  ```
  ```lang-fql {.result}
  /user(9323,"Timothy","Johnson",37)=nil
  /user(24335,"Andrew","Johnson",42)=nil
  /user(33423,"Ryan","Johnson",0x0ffa83,42.2)=nil
  ```
  FQL is an [open source](https://github.com/janderland/fdbq)
  query language for
  [Foundation DB](https://www.foundationdb.org/)
  defined via a simple, yet powerful
  [context-free
  grammar](https://github.com/janderland/fdbq/blob/main/syntax.ebnf).
...

# Overview

FQL queries look like key-values encoded using the directory
& tuple
[layers](https://apple.github.io/foundationdb/layer-concept.html).

```lang-fql {.query}
/my/directory("my","tuple")=4000
```

FQL queries may define a single key-value to be written, as
shown above, or may define a set of key-values to be read,
as shown below.

```lang-fql {.query}
/my/directory("my","tuple")=<int>
```
```lang-fql {.result}
/my/directory("my","tuple")=4000
```

The query above has a variable `<int>` as its value.
Variables act as placeholders for any of the supported [data
elements](#data-elements). This query will return a single
key-value from the database, if such a key exists.

FQL queries can also perform range reads & filtering by
including a variable in the key's tuple. The query below
will return all key-values which conform to the schema
defined by the query.

```lang-fql {.query}
/my/directory(<>,"tuple")=nil
```
```lang-fql {.result}
/my/directory("your","tuple")=nil
/my/directory(42,"tuple")=nil
```

All key-values with a certain key prefix can be range read
by ending the key's tuple with `...`.

```lang-fql {.query}
/my/directory("my","tuple",...)=<>
```
```lang-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

A query's value may be omitted to imply a variable, meaning
the following query is semantically identical to the one
above.

```lang-fql {.query}
/my/directory("my","tuple",...)
```
```lang-fql {.result}
/my/directory("my","tuple")=0x0fa0
/my/directory("my","tuple",47.3)=0x8f3a
/my/directory("my","tuple",false,0xff9a853c12)=nil
```

Including a variable in the directory tells FQL to perform
the read on all directory paths matching the schema.

```lang-fql {.query}
/<>/directory("my","tuple")
```
```lang-fql {.result}
/my/directory("my","tuple")=0x0fa0
/your/directory("my","tuple")=nil
```

Key-values can be cleared by using the special `clear` token
as the value.

```lang-fql {.query}
/my/directory("my","tuple")=clear
```

The directory layer can be queried by only including
a directory path.

```lang-fql {.query}
/my/<>
```
```lang-fql {.result}
/my/directory
```

# Data Elements

An FQL query contains instances of data elements. These are
the same types of elements found in the [tuple
layer](https://github.com/apple/foundationdb/blob/main/design/tuple.md).
Example instances of these elements can be seen below.

<div>

| Type     | Example                                |
|:---------|:---------------------------------------|
| `nil`    | `nil`                                  |
| `int`    | `-14`                                  |
| `uint`   | `7`                                    |
| `bool`   | `true`                                 |
| `float`  | `33.4`                                 |
| `bigint` | `#35299340192843523485929848293291842` |
| `string` | `"string"`                             |
| `bytes`  | `0xa2bff2438312aac032`                 |
| `uuid`   | `5a5ebefd-2193-47e2-8def-f464fc698e31` |
| `tuple`  | `("hello",27.4,nil)`                   |

</div>

> `bigint` support is not yet implemented.

Tuples & values may contain any of the data elements.

```lang-fql {.query}
/region/north_america(22.3,-8)=("rain","fog")
/region/east_asian("japan",nil)=0xff
```

Strings are the only data element allowed in directories. If
a directory string only contains alphanumericals,
underscores, dashes, and periods then the quotes don't need
to be included.

```lang-fql {.query}
/quoteless-string_in.dir(true)=false
/"other ch@r@cters must be quoted!"(20)=32.3
```

Quoted strings may contain quotes via backslash escapes.

```lang-fql {.query}
/my/dir("I said \"hello\"")=nil
```

# Value Encoding

The directory and tuple layers are responsible for encoding
the data elements in the key. As for the value, FDB doesn't
provide a standard encoding.

The table below outlines how FQL encodes data elements as
values. Endianness is configurable.

<div>

| Type     | Encoding                        |
|:---------|:--------------------------------|
| `nil`    | empty value                     |
| `int`    | 64-bit, 1's compliment          |
| `uint`   | 64-bit                          |
| `bool`   | single byte, `0x00` means false |
| `float`  | IEEE 754                        |
| `bigint` | not implemented yet             |
| `string` | ASCII                           |
| `bytes`  | as provided                     |
| `uuid`   | RFC 4122                        |
| `tuple`  | tuple layer                     |

</div>

# Variables

Variables allow FQL to describe key-value schemas. Any [data
element](#data-elements) may be replaced with a variable.
Variables are specified as a list of element types,
separated by `|`, wrapped in angled braces.

```lang-fql
<uint|string|uuid|bytes>
```

A variable may be empty, including no element types, meaning
it represents all element types.

```lang-fql
<>
```

Before the type list, a variable can be given a name. This
name is used to reference the variable in subsequent
queries, allowing for [index
redirection](#index-indirection).

```lang-fql {.query}
/index("cars",<varName:int>)
/data(:varName,...)
```
```lang-fql {.result}
/user(33,"mazda")=nil
/user(33,"ford")=nil
/user(33,"chevy")=nil
```

# Space & Comments

Whitespace and newlines are allowed within a tuple, between
its elements.

```lang-fql {.query}
/account/private(
  <uint>,
  <uint>,
  <string>,
)=<int>
```

Comments start with a `%` and continue until the end of the
line. They can be used to describe a tuple's elements.

```lang-fql
% private account balances
/account/private(
  <uint>,   % user ID
  <uint>,   % group ID
  <string>, % account name
)=<int>     % balance in USD
```

# Kinds of Queries

FQL queries can write & clear a single key-value, read one
or more key-values, or list directories. Throughout this
section, snippets of Go code are included to show how the
queries interact with the FDB API.

## Writes & Clears

Queries lacking a [variable](#variables) or the `...` token
perform mutations on the database by either writing
a key-value or clearing an existing one.

> Queries lacking a value imply an empty
> [variable](#variables) as the value and should not be
> confused with write queries.

If the query has a [data element](#data-elements) as its
value then it performs a write operation.

```lang-fql {.query}
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

Queries with the `clear` token as their value result in
a key-value being cleared.

```lang-fql {.query}
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

## Single Reads

If the query only has a [variable](#variables) or `...` in
its value (not its key) then it reads a single key-value,
if the key-value exists.

```lang-fql {.query}
/my/dir(99.8, 7dfb10d1-2493-4fb5-928e-889fdc6a7136)=<int|string>
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
cannot be decoded, the key-value does not matching the
schema.

If the value is specified as an empty variable, then the raw
bytes are returned.

```lang-fql {.query}
/some/data(10139)=<>
```

```lang-go {.equiv-go}
db.Transact(func(tr fdb.Transaction) (interface{}, error) {
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

## Range Reads

Queries with [variables](#variables) or the `...` token in
their key result in a range of key-values being read.

```lang-fql {.query}
/people(3392, <string|int>, <>)=(<uint>, ...)
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

The actual implementation of range-reads pipelines the
reading, filtering, and value decoding across multiple
threads.

# Filtering

Read queries define a schema to which key-values may or
may-not conform. In the Go snippets above, non-conformant
key-values were being filtered out of the results.

> Filtering is performed on the client-side and may result
> in lots of data being transferred to the host machine.
> Care must be taken to avoid wasted bandwidth.

Alternatively, FQL can throw an error when encountering
non-conformant key-values. This may help enforce the
assumption that all key-values within a directory conform to
the same schema.

# Index Indirection

TODO: Finish section.

```lang-fql {.query}
/user/index/surname("Johnson",<userID:int>)
/user/entry(:userID,...)
```

# Transaction Boundaries

TODO: Finish section.

# Design Recipes

TODO: Finish section.

# As a Layer

When integrating SQL into other languages, there are usually
two choices each with their own drawbacks:

1. Write literal _SQL strings_ into your code. This is
   simple but type safety isn't usually checked till
   runtime.

2. Use an _ORM_. This is more complex and sometimes doesn't
   perfectly model SQL semantics, but does provide type
   safety.

FQL leans towards option #2 by providing a Go API which is
structurally equivalent to the query language, allowing FQL
semantics to be modeled in the host language's type system.

This Go API may also be viewed as an FDB layer which unifies
the directory & tuple layers with the FDB base API.

```lang-go
package example

import (
  "github.com/apple/foundationdb/bindings/go/src/fdb"
  "github.com/apple/foundationdb/bindings/go/src/fdb/directory"

  "github.com/janderland/fdbq/engine"
  "github.com/janderland/fdbq/engine/facade"
  kv "github.com/janderland/fdbq/keyval"
)

func _() {
  fdb.MustAPIVersion(620)
  eg := engine.New(facade.NewTransactor(
    fdb.MustOpenDefault(), directory.Root()))

  // /user/entry(22573,"Goodwin","Samuels")=nil
  query := kv.KeyValue{
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

  // Perform the write.
  err := eg.Set(query);
  if err != nil {
    panic(err)
  }
}
```
