# FDBQ

FDBQ provides a query language and an alternative client API for Foundation DB.
Some things this project aims to achieve are:

- [x] Provide a textual description of key-value schemas.
- [x] Provide an intuitive query language for FDB.
- [x] Provide a Go API which is structurally equivalent to the query language.
- [ ] Improve the ergonomics of the FoundationDB API.
    - [ ] Gracefully handle multi-transaction range-reads.
    - [ ] Gracefully handle transient errors.
- [ ] Standardize the encoding of primitives (int, float, bool) as FDB values.

## Building & Running

FDBQ is available as a Docker image for running queries. The first argument
passed to the container is the contents of the cluster file. The remaining
arguments are passed to the FDBQ binary.

```bash
# 'my_cluster:baoeA32@172.20.3.33:4500' is used as the contents
# for the cluster file. '-log' and '/my/dir{<>}=42' are passed
# as args to the FDBQ binary.
docker run docker.io/janderland/fdbq 'my_cluster:baoeA32@172.20.3.33:4500' -log '/my/dir{<>}=42'
```

The cluster file contents (first argument) is evaluated by Bash within the
container before being written to disk, which allows for converting hostnames
into IPs.

```bash
# The cluster file contents includes a bit of Bash which
# converts the hostname 'fdb' to an IP address before
# writing the cluster file on to the container's disk.
CFILE='docker:docker@$(getent hosts fdb | cut -d" " -f1):4500'
docker run docker.io/janderland/fdbq $CFILE -log '/my/dir{<>}=42'
```

## Query Language

Here is the [syntax definition](syntax.ebnf) for the query language. Currently,
FDBQ is focused on reading & writing key-values created using the directory and
tuple layers. Reading or writing keys of arbitrary byte strings is not
supported.

FDBQ queries are a textual representation of a specific key-value or a schema
describing the structure of many key-values. These queries have the ability to
write a key-value, read one or more key-values, and list directories.

### Language Components

#### Primitives

FDBQ utilizes textual representations of the element types supported by the
tuple layer. These are known as primitives. Besides as tuple elements,
primitives can also be used as the value portion of a key-value.

```fdbq
17
-23
33.4
nil
true
false
"string"
0xa2bff2438312aac032
5a5ebefd-2193-47e2-8def-f464fc698e31
```

#### Directories

A directory is specified as a sequence of strings, each prefixed by a forward
slash:

```fdbq
/my/dir/path_way
```

The strings of the directory do not need quotes if they only contain
alphanumerical, underscores, dashes, or periods. To use other symbols, the
strings must be quoted:

```
/my/"dir@--\o/"/path_way
```

The quote character may be backslash escaped:

```
/my/"\"dir\""/path_way
```

#### Tuples

A tuple is specified as a sequence of elements, separated by commas, wrapped in
a pair of curly braces. The elements may be a tuple or any of the primitive
types.

```fdbq
{"one", 2, 0x03, { "subtuple" }, 5825d3f8-de5b-40c6-ac32-47ea8b98f7b4}
```

Any combination of spaces, tabs, and newlines is allowed after the opening  
brace and commas.

```fdbq
{
  1,
  2,
  3,
}
```

#### Key-Values

A key-value is specified as a directory, tuple, '=' symbol, and value appended
together:

```fdbq
/my/dir{"this", 0}=0xabcf03
```

The value following the '=' symbol may be any of the primitives or a tuple:

```fdbq
/my/dir{22.3, -8}={"another", "tuple"}
```

### Kinds of Queries

The following examples showcase FDBQ queries and the equivalent FDB API calls
implemented in Go.

#### Set

```fdbq
/my/dir{"hello", "world"}=42
```

```go
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

#### Clear

```fdbq
/my/dir{"hello", "world"}=clear
```

```go
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

#### Get Single Key

```fdbq
/my/dir{99.8, 7dfb10d1-2493-4fb5-928e-889fdc6a7136}
```

```go
db.Transact(func(tr fdb.Transaction) (interface{}, error) {
  dir, err := directory.Open(tr, []string{"my", "dir"}, nil)
  if err != nil {
    if errors.Is(err, directory.ErrDirNotExists) {
      return nil, nil
    }
    return nil, err
  }

  return tr.Get(dir.Pack(tuple.Tuple{99.8,
    tuple.UUID{0x7d, 0xfb, 0x10, 0xd1, 0x24, 0x93, 0x4f, 0xb5, 0x92, 0x8e, 0x88, 0x9f, 0xdc, 0x6a, 0x71, 0x36}))
})
```

#### Read Range of Keys

```fdbq
/people{3392, <string|int>, <>}={<uint>, ...}
```

```go
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
      return nil, fmt.Errorf("invalid kv: %v", kv)
    }

    switch tup[0].(type) {
    default:
      return nil, fmt.Errorf("invalid kv: %v", kv)
    case string | int64:
    }

    val, err := tuple.Unpack(kv.Value)
    if err != nil {
      return nil, fmt.Errorf("invalid kv: %v", kv)
    }
    if len(val) == 0 {
      return nil, fmt.Errorf("invalid kv: %v", kv)
    }
    if _, isInt := val[0].(uint64); !isInt {
      return nil, fmt.Errorf("invalid kv: %v", kv)
    }

    results = append(results, kv)
  }
  return results, nil
})
```

#### Read & Filter Directory Paths

```fdbq
/root/<>/items/<>
```

```go
db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
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
})
```
