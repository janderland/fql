

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
| `be`       | Big endian      |
| `le`       | Little Endian   |
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
| `be`        | Big endian    |
| `le`        | Little Endian |
| `f32`       | 32-bit        |
| `f64`       | 64-bit        |
| `f80`       | 80-bit        |

</div>

