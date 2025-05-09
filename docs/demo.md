# FQL

## How is FDB different?

- Distributed key-value DB with ACID transactions
- Atomicity, isolation, & durability handled by DB
- Consistency distributed accross clients
- Ordered keys make range-reads efficient
- Lacks a query language

## How is FQL different?

- Designed around DB's semantics
- Core data structure is a key-value
- Composable and intuitive

## Write Query

```language-fql {.query}
/my/dir("hello", "world")=42
```

```lang-go {.equiv-go}
db.Transact(func(tr fdb.Transaction) (any, error) {
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

## Read Query

```language-fql {.query}
/my/dir(99.8, 7dfb10d1-2493-4fb5-928e-889fdc6a7136)=<int|str>
```

```lang-go {.equiv-go}
db.Transact(func(tr fdb.Transaction) (any, error) {
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
  
     
  if len(val) == 8 {
      return binary.LittleEndian.Uint64(val), nil
  }
  return string(val), nil
})
```

## Read With Filtering

```language-fql {.query}
/people(3392, <str|int>, <>)=(<uint>, ...)
```

```lang-go {.equiv-go}
db.ReadTransact(func(tr fdb.ReadTransaction) (any, error) {
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
