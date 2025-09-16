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

FQL is specified as a context-free [grammar][]. The queries
look like key-values encoded using the [directory][] and
[tuple][] layers. To the left of the `=` is the key which
includes a directory path and tuple. To the right is the
value.

[grammar]: https://github.com/janderland/fql/blob/main/syntax.ebnf
[directory]: https://apple.github.io/foundationdb/developer-guide.html#directories
[tuple]: https://apple.github.io/foundationdb/data-modeling.html#data-modeling-tuples

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
by including a variable in the key's tuple. The query below
will return all key-values which conform to the schema
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
by ending the key's tuple with `...`.

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

# Data Elements

An FQL query contains instances of data elements. These
mirror the types of elements found in the [tuple layer][].
This section will describe how data elements behave in the
FQL language, while [element encoding](#element-encoding)
describes how FQL encodes the elements before writing them
to the DB.

[tuple layer]: https://github.com/apple/foundationdb/blob/main/design/tuple.md

<div>

| Type    | Description    | Examples                               |
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
large integer. For example, the integer in the query below
doesn't fit in a 64-bit value.

```
/bigint(92233720368547758084)=nil
```

The `num` type may be instantiated as any real number which
can be approximated by an [80-bit floating point][] value,
in accordance with IEEE 754. The implementation determines
the exact range of allowed values. Scientific notation may
be used. The type may also be instantiated as the tokens
`-inf`, `inf`, `-nan`, or `nan`. 

[80-bit floating point]: https://en.wikipedia.org/wiki/Extended_precision#x86_extended_precision_format

```language-fql {.query}
/float(-inf,nan)=1.234e4732
```

The `str` type may be instantiated as a unicode string
wrapped in double quotes. It is the only element type
allowed in directory paths. If a directory string only
contains alphanumericals, underscores, dashes, and periods
then the quotes may be excluded. Quoted strings may contain
double quotes via backslash escapes.

```language-fql {.query}
/quoteless-string_in.dir("escape \"wow\"")=nil
/"other ch@r@cters must be 'quoted'"(nil)=""
```

The `uuid` and `bytes` types may be instantiated using
upper, lower, or mixed case hexidecimal numbers. For `uuid`,
the numbers are grouped in the standard 8, 4, 4, 4, 12
format. For `bytes`, any even number of hexidecimal digits
are prefixed by `0x`.

```language-fql {.query}
/hex(fC2Af671-a248-4AD6-ad57-219cd8a9f734)=0x3b42ADED28b9
```

The `tup` type may contain any of the data elements,
including sub-tuples. 

```language-fql {.query}
/sub/tuple("japan",("sub",nil))=0xff
/tuple/value(22.3,-8)=("rain","fog")
```

# Holes & Schemas

Holes are a group of syntax constructs used to define
a key-value schema by acting as placeholders for one or more
data elements. There are three kinds of holes: variables,
references, and the `...` token.

Variables are used to represent a single [data
element](#data-elements). Variables are specified as a list
of element types, separated by `|`, wrapped in angled
braces.

```language-fql
<int|str|uuid|bytes>
```

The variable's type list describes which kinds of data
elements are allowed at the variable's position. A variable
may be empty, including no element types, meaning it
represents all element types.

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
/index("cars",<varName:int>)
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

Options modify the semantics of [data
elements](#data-elements), [variables](#holes-schemas), and
[queries](#basic-queries). They can instruct FQL to use
alternative encodings, limit a query's result count, or
change other behaviors.

Options are specified as a comma separated list wrapped in
braces. For instance, to specify that an `int` should be
encoded as a little-endian unsigned 8-bit integer, the
following options would be included after the element.

```language-fql
3548[u8,le]
```

Similarly, if a variable should only match against
a big-endian 32-bit float then the following option would be
included after the `num` type.

```language-fql
<num[f32,be]>
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

# Element Encoding

FoundationDB stores the keys and values as simple byte
strings leaving the client responsible for encoding the
data. FQL determines how to encode [data
elements](#data-elements) based on their data type, position
within the query, and associated [options](#options).

Keys are *always* encoded using the [directory][] and
[tuple][] layers. Write queries create directories if they
do not exist.

[directory]: https://apple.github.io/foundationdb/developer-guide.html#directories
[tuple]: https://apple.github.io/foundationdb/data-modeling.html#data-modeling-tuples

```language-fql {.query}
/directory/"p@th"(nil,57223,0xa8ff03)=nil
```

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  dir, err := CreateOrOpenDir(tr, []string{"directory", "p@th"})
  if err != nil {
    return nil, err
  }

  // Pack the tuple and prepend the dir prefix
  key := dir.Pack(Tuple{nil, 57223, []byte{0xa8, 0xff, 0x03}})

  tr.Set(key, nil)
  return nil, nil
})
```

If a query reads from a directory which doesn't exist,
nothing is returned. The tuple layer encodes metadata about
element types, allowing FQL to decode keys without a schema.

```language-fql {.query}
/directory/<>(...)
```

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  dir, err := OpenDir(tr, []string{"directory"})
  if err != nil {
    if err == DirNotExists {
      return nil, nil
    }
    return nil, err
  }

  // List the sub-directories of dir
  subDirs, err := dir.List(tr)
  if err != nil {
    return nil, err
  }

  // For each sub-directory, grab all the KVs
  var results []KeyValue
  for _, subDir := range subDirs {
    iter := tr.GetRange(subDir).Iterator()

    for iter.Advance() {
      kv := iter.MustGet()

      // Remove the directory prefix and unpack the tuple
      tup, err := dir.Unpack(kv.Key)
      if err != nil {
        // Return partial results with error
        return results, err
      }

      val := // Value unpacking will be discussed later...

      results = append(results, KeyValue{
        Key: Key{
          Directory: dir,
          Tuple: tup,
        },
        Value: val,
      })
    }
  }
  return results, nil
})
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

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  key := // Encode the key...

  val, err := PackTup(Tuple{42})
  if err != nil {
    return nil, err
  }

  tr.Set(key, val)
  return nil, nil
})
```

This default encoding allows values to be decoded without
knowing their type.

```language-fql {.query} 
/people/age("jon","smith")=<>
```

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  key := // Encode the key...

  valBytes := tr.MustGet(key)

  // Assume the value is a tuple
  valTup, err := UnpackTup(valBytes)
  if err == nil {
    return tup, nil
  }

  // The value isn't a tuple, so return raw bytes
  return valBytes, err
})
```

Using options, values can be encoded in other ways. For
instance, the option `u16` tells FQL to encode an unsigned
integer using 16-bits. The byte order can be specified using
the options `le` and `be` for little and big endian
respectively. 

```language-fql {.query}
/numbers/big("37")=37[i16,be]
```

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  key := // Encode the key...

  // Pack the value into 16 bits.
  val := make([]byte, 2)
  binary.BigEndian.PutUint64(val, 37)

  tr.Set(key, val)
  return nil, nil
})
```

If the value was encoded with non-default values, then the
encoding must be specified in the variable. 

```language-fql {.query}
/numbers/big("37")=<int[i16,be]>
```

```lang-go {.equiv-go}
db.Transact(func(tr Transaction) (any, error) {
  key := // Encode the key...

  valBytes := tr.MustGet(key)

  // Unpack bytes as a 16-bit int
  valUnsigned := binary.BigEndian.Uint16(valBytes)
  val := int16(valUnsigned)

  return val, err
})
```

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

<!-- vim: set tw=60 :-->
