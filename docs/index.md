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
  Foundation QL is a query language for [Foundation
  DB](https://www.foundationdb.org/). FQL aims to make FDB's
  semantics feel natural and intuitive.
...

## TODO

- Only allow `...` in the key's root tuple.
- Better descriptions for data elements.

## Overview

FQL queries generally look like key-values. They have a key
(directory & tuple) and value separated by `=`. FQL can only
access keys encoded using the directory & tuple
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

The query above has a variable `<int>` as it's value.
Variables act as placeholders for any of the supported [data
elements](#data-elements). This query will return a single
key-value from the database, if such a key exists.

FQL queries can also perform range reads by including
a variable in the key's tuple. The query below will return
all key-values which conform to the schema defined by the
query. 

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

The next two sections of this document elaborate on the
language's grammar and semantics. If you wish to immediately
see more examples of the language in practice, skip to
[design recipes](#design-recipes).

## Grammar

FQL is a context-free language with a formal
[definition](https://github.com/janderland/fdbq/blob/main/syntax.ebnf).
This section elaborates on this definition.

### Key-Values

Most FQL queries are structured like key-values and are
written as a [directory](#Directory), [tuple](#Tuple), `=`,
and value appended together.

```lang-fql
/app/data("server A",0)=0xabcf03
```

The value following the `=` may be any of the [data
elements](#data-elements) or a [variable](#variables).

```lang-fql
/region/north_america(22.3,-8)=("rain","fog")
/region/north_america(22.3,-8)=<tuple|int>
/region/north_america(22.3,-8)=-16
```

The value may also be the `clear` token.

```lang-fql
/some/where("home","town",88.3)=clear
```

### Directories

A directory is specified as a sequence of strings, each
prefixed by a forward slash:

```lang-fql
/my/dir/path_way
```

The strings of the directory do not need quotes if they only
contain alphanumericals, underscores, dashes, or periods. To
use other symbols, the strings must be quoted:

```lang-fql
/my/"dir@--o/"/path_way
```

The quote character may be backslash escaped:

```lang-fql
/my/"\"dir\""/path_way
```

### Tuples

A tuple is specified as a sequence of [data
elements](#data-elements) and [variables](#variables),
separated by commas, wrapped in a pair of parenthesis.
Sub-tuples are allowed.

```lang-fql
("one",0x03,("subtuple"),5825d3f8-de5b-40c6-ac32-47ea8b98f7b4)
```

The last element of a tuple may be the `...` token.

```lang-fql
(0xff,"thing",...)
```

Any combination of spaces, tabs, and newlines are allowed
after the opening brace and commas. Trailing commas are
allowed.

```lang-fql
(
  1,
  2,
  3,
)
```

### Data Elements

In a FQL query, the directory, tuples, and value contain
instances of data elements. FQL utilizes the same types of
elements as the [tuple
layer](https://github.com/apple/foundationdb/blob/main/design/tuple.md).
Example instances of these types can be seen below.

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

### Variables

Any [data element](#data-elements) may be replaced with
a variable. Variables are specified as a list of data types,
separated by `|`, wrapped in angled braces.

```lang-fql
<uint|string|uuid|bytes>
```

A variable may be empty, including no data types.

```lang-fql
<>
```

### Comments

Comments start with `%` and continue until the end of the
line.

```lang-fql
% This query will read all the first
% names. A single name may be returned
% multiple times.

/index/name(<name:string>,...)
```

You can add comments within a tuple or after the value to
describe the data elements.

```lang-fql
/account/private(
  <uint>,   % user ID
  <uint>,   % group ID
  <string>, % account name
)=<int>     % balance in USD
```

## Semantics

FQL queries have the ability to write a single key-value,
clear a single key-value, read one or more key-values, and
list directories. This section elaborates on what queries do
and how they encode/decode data.

Throughout this section, snippets of Go code are included to
help explemplify what's being discussed. These snippets
accurately showcases the DB operations in the clearest way
possible and don't include the optimizations and concurrency
implemented in the FQL engine.

### Writes

Queries lacking a [variable](#variables) or the `...` token
perform mutations on the database by either writing
a key-value or clearing on existing one. Queries lacking
a value imply an empty [variable](#variables) as the value
and should not be confused with write queries.

If the query has a [data element](#data-elements) as it's
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

### Reads

If the query contains a [variable](#variables) or `...`
token, then it performs a read. Queries lacking a value
imply an empty [variable](#variables) as their value and are
therefore read queries.

If the query lacks a [variable](#variables) or `...` in it's
key then it reads a single-value, if the key-value exists.

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
    tuple.UUID{0x7d, 0xfb, 0x10, 0xd1, 0x24, 0x93, 0x4f, 0xb5, 0x92, 0x8e, 0x88, 0x9f, 0xdc, 0x6a, 0x71, 0x36}))
  
     
  if len(val) == 8 {
      return binary.LittleEndian.Uint64(val), nil
  }
  return string(val), nil
})
```

Queries with [variables](#variables) or a `...` token in
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

    if len(tup) != 3 {
      continue
    }

    switch tup[0].(type) {
    default:
      continue
    case string | int64:
    }

    val, err := tuple.Unpack(kv.Value)
    if err != nil {
      continue
    }
    if len(val) == 0 {
      continue
    }
    if _, isInt := val[0].(uint64); !isInt {
      continue
    }

    results = append(results, kv)
  }
  return results, nil
})
```

Read queries define a schema to which key-values may or
may-not conform. In the Go snippet above, you may have
noticed that non-conformant key-values are being filtered
out of the results.

Alternatively, FQL can throw an error when encountering
a non-conformant key-value. This may help enforce the
assumption that all key-values within a directory conform to
the same schema. This behavior, and others, can be
configured via the transaction's [options](#options).

### Data Encoding

The directory and tuple layers are responsible for encoding
the data elements in the key. As for the value, FDB doesn't
provide a standard encoding.

The table below outlines how data elements are encoded 
when present in the value section.

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

### Index Indirection

TODO: Finish section.

```lang-fql {.query}
/user/index/surname("Johnson",<userID:int>)
/user/entry(:userID,...)
```

### Transaction Boundaries

TODO: Finish section.

## Design Recipes

TODO: Finish section.

## As a Layer

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

