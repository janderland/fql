# TODO: Rust Implementation Improvements

This document outlines potential improvements to make the FQL Rust implementation more idiomatic, performant, and maintainable.

## 🎯 High Priority - Idiomatic Rust

### Type Safety & Error Handling

- [ ] **Replace `String` errors with proper error types**
  - `convert.rs`: Use `thiserror` for `ConversionError` instead of `Result<_, String>`
  - Implement `std::error::Error` for all error types
  - Add context with error source chains

- [ ] **Implement standard traits across the board**
  - `Display` for all query types (for debugging and logging)
  - `FromStr` for parsing from strings (complement to `format`)
  - `TryFrom`/`TryInto` for conversions between types
  - `Default` where appropriate (e.g., `EngineConfig`)
  - `Hash` for types that should be hashable

- [ ] **Add newtype wrappers for semantic clarity**
  ```rust
  // Instead of raw Vec<u8>
  pub struct PackedValue(Vec<u8>);
  pub struct PackedTuple(Vec<u8>);

  // Prevents mixing up different byte representations
  impl PackedValue {
      pub fn as_bytes(&self) -> &[u8] { &self.0 }
  }
  ```

### API Ergonomics

- [ ] **Builder pattern for complex types**
  ```rust
  // Instead of verbose construction
  KeyValue::builder()
      .directory(["users", "profiles"])
      .tuple([TupElement::Int(42)])
      .value(Value::String("data"))
      .build()
  ```

- [ ] **Convenience constructors with `Into<T>`**
  ```rust
  impl Key {
      pub fn new(dir: impl Into<Directory>, tup: impl Into<Tuple>) -> Self {
          Self { directory: dir.into(), tuple: tup.into() }
      }
  }

  // Allows: Key::new(vec!["users"], vec![42])
  // Via From<Vec<&str>> for Directory and From<Vec<i64>> for Tuple
  ```

- [ ] **Method chaining for queries**
  ```rust
  let query = Query::new()
      .in_directory(["users"])
      .with_tuple([42, "key"])
      .equals(Value::String("value"));
  ```

- [ ] **Add `as_ref()` and `as_mut()` accessors**
  - Reduces need for pattern matching when you just need inner data
  - Implement for `Query`, `Value`, `TupElement`, `DirElement`

### Zero-Copy & Performance

- [ ] **Use `Cow<'a, str>` for strings where appropriate**
  ```rust
  pub enum Value<'a> {
      String(Cow<'a, str>),  // Can borrow or own
      // ... other variants
  }
  ```

- [ ] **Add `&[u8]` variants alongside `Vec<u8>`**
  ```rust
  pub fn pack_into(&self, buf: &mut Vec<u8>) -> Result<(), SerializationError>;
  pub fn pack_bytes(&self) -> Result<&[u8], SerializationError>;
  ```

- [ ] **Use `SmallVec` for tuples and directories**
  - Most tuples/directories are small (< 8 elements)
  - Avoids heap allocation for common case
  ```rust
  pub type Tuple = SmallVec<[TupElement; 8]>;
  pub type Directory = SmallVec<[DirElement; 4]>;
  ```

- [ ] **Implement `serde` `zero-copy` deserialization**
  - Use `#[serde(borrow)]` where possible
  - Reduces allocations when deserializing

## 🔧 Medium Priority - Robustness

### Validation & Constraints

- [ ] **Add validation functions**
  ```rust
  impl KeyValue {
      pub fn validate(&self) -> Result<(), ValidationError> {
          // Check for Variables in constant values
          // Check MaybeMore only at end of tuple
          // Check no nil values where not allowed
      }
  }
  ```

- [ ] **Use typed builders with compile-time validation**
  ```rust
  // Type state pattern to prevent invalid queries
  pub struct QueryBuilder<State> {
      _state: PhantomData<State>,
      // ...
  }

  impl QueryBuilder<NeedsValue> {
      pub fn value(self, v: Value) -> QueryBuilder<Complete> { ... }
  }
  ```

- [ ] **Add `#[non_exhaustive]` to enums for future compatibility**
  ```rust
  #[non_exhaustive]
  pub enum ValueType { ... }
  ```

### Testing & Documentation

- [ ] **Add property-based tests with `proptest`**
  ```rust
  proptest! {
      #[test]
      fn pack_unpack_roundtrip(value: Value) {
          let packed = pack(&value, Endianness::Big, false)?;
          let unpacked = unpack(&packed, infer_type(&value), Endianness::Big)?;
          assert_eq!(value, unpacked);
      }
  }
  ```

- [ ] **Add doc tests to all public functions**
  - Current code has some doc comments but no testable examples
  - Doc tests serve as both documentation and regression tests

- [ ] **Add benchmark suite**
  ```rust
  #[bench]
  fn bench_tuple_comparison(b: &mut Bencher) {
      let schema = /* ... */;
      let candidate = /* ... */;
      b.iter(|| compare(&schema, &candidate));
  }
  ```

- [ ] **Add integration tests**
  - Test cross-crate interactions
  - Test parsing → classification → execution flow
  - Test format → parse roundtrips

### Error Context

- [ ] **Add rich error context with `miette`**
  ```rust
  #[derive(Error, Diagnostic)]
  #[error("Tuple comparison failed")]
  pub struct ComparisonError {
      #[source_code]
      src: String,
      #[label("mismatch at this position")]
      span: SourceSpan,
  }
  ```

- [ ] **Add error recovery hints**
  ```rust
  #[error("Cannot serialize Variable")]
  #[help("Variables are placeholders for read queries. Use a concrete value for write operations.")]
  CannotSerialize(String),
  ```

## 🚀 Low Priority - Advanced Features

### Type System Enhancements

- [ ] **Use const generics for fixed-size arrays**
  ```rust
  pub struct VStamp<const N: usize = 10> {
      pub tx_version: [u8; N],
      pub user_version: u16,
  }
  ```

- [ ] **Phantom types for query classification**
  ```rust
  pub struct Query<Class = Unclassified> {
      inner: QueryInner,
      _class: PhantomData<Class>,
  }

  pub struct Constant;
  pub struct ReadRange;
  // etc.

  // Type-safe API that knows query class at compile time
  impl Query<ReadRange> {
      pub async fn execute_range(&self) -> Result<Vec<KeyValue>> { ... }
  }
  ```

- [ ] **Sealed traits for closed hierarchies**
  ```rust
  mod private {
      pub trait Sealed {}
  }

  pub trait TupleElement: private::Sealed {
      fn as_bytes(&self) -> Vec<u8>;
  }

  // Prevents external crates from implementing TupleElement
  ```

### Async/Await Improvements

- [ ] **Make Engine use `async_trait` properly**
  ```rust
  #[async_trait]
  pub trait QueryExecutor {
      async fn execute(&self, query: &Query) -> Result<Vec<KeyValue>>;
  }
  ```

- [ ] **Add `Stream` support for range queries**
  ```rust
  pub async fn read_range_stream(&self, kv: &KeyValue)
      -> Result<impl Stream<Item = Result<KeyValue>>> {
      // Proper async streaming instead of collecting to Vec
  }
  ```

- [ ] **Add cancellation support with `tokio::select!`**
  ```rust
  pub async fn execute_with_timeout(
      &self,
      query: &Query,
      timeout: Duration
  ) -> Result<Vec<KeyValue>> {
      tokio::select! {
          result = self.execute(query) => result,
          _ = tokio::time::sleep(timeout) => Err(EngineError::Timeout),
      }
  }
  ```

### Parser Improvements

- [ ] **Implement full parser state machine**
  - Currently just a placeholder
  - Port the Go parser's state machine to Rust
  - Use `nom` or `pest` for robust parsing

- [ ] **Add parser error recovery**
  ```rust
  pub struct ParseError {
      pub position: usize,
      pub expected: Vec<&'static str>,
      pub found: String,
  }
  ```

- [ ] **Support parser streaming**
  - Parse from `Read` trait instead of just `&str`
  - Useful for large queries or stdin

### Serialization Improvements

- [ ] **Add compression support**
  ```rust
  pub fn pack_compressed(
      value: &Value,
      compression: Compression,
  ) -> Result<Vec<u8>> { ... }
  ```

- [ ] **Add custom serialization formats**
  - JSON (via serde_json)
  - MessagePack
  - CBOR
  - Allows debugging and cross-language interop

- [ ] **Implement FoundationDB tuple packing**
  - Currently placeholder in `convert.rs` and `values.rs`
  - Integrate with `foundationdb` crate
  - Proper tuple layer implementation

## 📚 Documentation Improvements

- [ ] **Add architectural decision records (ADRs)**
  - Document why enums over visitor pattern
  - Document async/await choices
  - Document error handling strategy

- [ ] **Add comprehensive examples**
  ```rust
  //! # Examples
  //!
  //! ## Basic query construction
  //! ```
  //! use fql::*;
  //!
  //! let query = Query::key_value(
  //!     Key::new(vec![], vec![TupElement::Int(42)]),
  //!     Value::String("hello".into())
  //! );
  //! ```
  ```

- [ ] **Add migration guide from Go**
  - Show Go code → Rust equivalent
  - Highlight idiom differences
  - Performance comparison

- [ ] **Add performance tuning guide**
  - When to use `&str` vs `String`
  - When to use `Cow`
  - Allocation patterns to avoid

## 🔐 Safety & Security

- [ ] **Add fuzzing targets**
  ```rust
  #[cfg(fuzzing)]
  pub fn fuzz_parse(data: &[u8]) {
      if let Ok(s) = std::str::from_utf8(data) {
          let _ = parser::parse(s);
      }
  }
  ```

- [ ] **Add sanitization for untrusted input**
  - Limit tuple depth
  - Limit string lengths
  - Prevent DOS via large allocations

- [ ] **Add `#[must_use]` to important types**
  ```rust
  #[must_use = "Query results should be checked"]
  pub async fn execute(&self, query: &Query) -> Result<Vec<KeyValue>>
  ```

## 🎨 Code Organization

- [ ] **Split large files into modules**
  - `values.rs` is 379 lines, split into `pack.rs` and `unpack.rs`
  - `tuple.rs` is 291 lines, split into `compare.rs` and `matches.rs`

- [ ] **Add prelude module**
  ```rust
  // keyval/src/prelude.rs
  pub use crate::{
      Query, KeyValue, Key, Value, TupElement, Variable,
      // Common traits
  };
  ```

- [ ] **Use workspace-level Clippy configuration**
  ```toml
  # .cargo/config.toml
  [target.'cfg(all())']
  rustflags = [
      "-W", "clippy::pedantic",
      "-W", "clippy::nursery",
  ]
  ```

## 🔄 API Compatibility

- [ ] **Add feature flags for optional functionality**
  ```toml
  [features]
  default = ["parser", "engine"]
  parser = ["nom"]
  engine = ["foundationdb", "tokio"]
  serde = ["serde", "serde_json"]
  ```

- [ ] **Version compatibility**
  - Document MSRV (Minimum Supported Rust Version)
  - Add CI to test against MSRV
  - Use `cargo-msrv` to track

## 🧪 Testing Infrastructure

- [ ] **Add mutation testing with `cargo-mutants`**
  - Ensures tests actually catch bugs
  - Finds untested code paths

- [ ] **Add coverage tracking**
  - Use `cargo-tarpaulin` or `cargo-llvm-cov`
  - Track coverage over time
  - Set minimum coverage thresholds

- [ ] **Add test fixtures**
  ```rust
  // tests/fixtures/mod.rs
  pub fn sample_queries() -> Vec<Query> { ... }
  pub fn sample_keyvalues() -> Vec<KeyValue> { ... }
  ```

## 🌍 Ecosystem Integration

- [ ] **Implement `slog` or `tracing` throughout**
  - Add structured logging to engine
  - Add debug traces to parser
  - Performance instrumentation

- [ ] **Add `clap` integration for CLI**
  - Currently basic, could be more ergonomic
  - Add shell completion
  - Add man page generation

- [ ] **Add `ratatui` TUI implementation**
  - Currently placeholder
  - Port Go Bubble Tea implementation
  - Add vim-like key bindings

---

## Priority Summary

**Start Here (High Priority):**
1. Implement standard traits (Display, FromStr, TryFrom, etc.)
2. Replace String errors with proper error types
3. Add builder patterns for ergonomic construction
4. Add comprehensive doc tests

**Next Steps (Medium Priority):**
1. Add validation functions
2. Implement property-based tests
3. Add rich error context
4. Complete parser implementation

**Future Work (Low Priority):**
1. Advanced type system features (const generics, phantom types)
2. Performance optimizations (SmallVec, zero-copy)
3. Enhanced async support (streaming, cancellation)
4. Fuzzing and security hardening
