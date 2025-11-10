//! Core data structures for FQL queries and key-values.
//!
//! This crate contains types representing key-values and related utilities.
//! These types model both queries and the data returned by queries. They can
//! be constructed from query strings using the parser crate, but are also
//! designed to be easily constructed directly in Rust source code.
//!
//! # Overview
//!
//! Unlike the Go implementation which uses the visitor pattern to work around
//! the lack of tagged unions, this Rust implementation uses enums directly,
//! providing type-safe polymorphism with pattern matching.

use serde::{Deserialize, Serialize};
use std::fmt;

pub mod class;
pub mod convert;
pub mod tuple;
pub mod values;

/// A query that can be passed to the engine. This includes KeyValue, Key, and Directory queries.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum Query {
    /// A key-value pair query (read or write)
    KeyValue(KeyValue),
    /// A key-only query (equivalent to KeyValue with empty Variable)
    Key(Key),
    /// A directory listing query
    Directory(Directory),
}

/// A key-value pair that can be passed as a query or returned as a result.
/// When returned as a result, it will not contain Variable, Clear, or MaybeMore.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct KeyValue {
    pub key: Key,
    pub value: Value,
}

/// A key consisting of a directory path and tuple.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Key {
    pub directory: Directory,
    pub tuple: Tuple,
}

/// A directory path consisting of elements that can be strings or variables.
pub type Directory = Vec<DirElement>;

/// An element in a directory path - either a concrete string or a variable placeholder.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum DirElement {
    String(String),
    Variable(Variable),
}

/// A tuple of elements. In FQL, tuples are the basic building blocks of keys.
pub type Tuple = Vec<TupElement>;

/// An element that can appear in a tuple.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum TupElement {
    /// A nested tuple
    Tuple(Tuple),
    /// Nil/empty element
    Nil,
    /// Signed 64-bit integer
    Int(i64),
    /// Unsigned 64-bit integer
    Uint(u64),
    /// Boolean value
    Bool(bool),
    /// 64-bit floating point
    Float(f64),
    /// UTF-8 string
    String(String),
    /// UUID (16 bytes)
    Uuid(uuid::Uuid),
    /// Arbitrary bytes
    #[serde(with = "serde_bytes")]
    Bytes(Vec<u8>),
    /// Variable placeholder with type constraints
    Variable(Variable),
    /// Special marker allowing additional tuple elements (only valid as last element)
    MaybeMore,
    /// Versionstamp with transaction and user versions
    VStamp(VStamp),
    /// Future versionstamp (assigned at commit time)
    VStampFuture(VStampFuture),
}

/// A value that can be stored in a key-value pair.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum Value {
    /// A tuple value
    Tuple(Tuple),
    /// Nil/empty value
    Nil,
    /// Signed 64-bit integer
    Int(i64),
    /// Unsigned 64-bit integer
    Uint(u64),
    /// Boolean value
    Bool(bool),
    /// 64-bit floating point
    Float(f64),
    /// UTF-8 string
    String(String),
    /// UUID (16 bytes)
    Uuid(uuid::Uuid),
    /// Arbitrary bytes
    #[serde(with = "serde_bytes")]
    Bytes(Vec<u8>),
    /// Variable placeholder with type constraints
    Variable(Variable),
    /// Clear marker (delete this key)
    Clear,
    /// Versionstamp with transaction and user versions
    VStamp(VStamp),
    /// Future versionstamp (assigned at commit time)
    VStampFuture(VStampFuture),
}

/// A variable placeholder defining a schema. Variables can have type constraints.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize, Default)]
pub struct Variable {
    /// Type constraints for this variable. Empty means any type is allowed.
    pub types: Vec<ValueType>,
}

impl Variable {
    /// Create a new variable that accepts any type.
    pub fn any() -> Self {
        Self::default()
    }

    /// Create a new variable with specific type constraints.
    pub fn with_types(types: Vec<ValueType>) -> Self {
        Self { types }
    }
}

// Builder patterns for ergonomic construction

/// Builder for constructing KeyValue queries.
#[derive(Debug, Clone, Default)]
pub struct KeyValueBuilder {
    directory: Directory,
    tuple: Tuple,
    value: Option<Value>,
}

impl KeyValueBuilder {
    /// Create a new KeyValueBuilder.
    pub fn new() -> Self {
        Self::default()
    }

    /// Set the directory path.
    pub fn directory<D: Into<Directory>>(mut self, dir: D) -> Self {
        self.directory = dir.into();
        self
    }

    /// Add a directory element.
    pub fn dir_element(mut self, elem: DirElement) -> Self {
        self.directory.push(elem);
        self
    }

    /// Add a string directory element.
    pub fn dir_str<S: Into<String>>(mut self, s: S) -> Self {
        self.directory.push(DirElement::String(s.into()));
        self
    }

    /// Set the tuple.
    pub fn tuple<T: Into<Tuple>>(mut self, tup: T) -> Self {
        self.tuple = tup.into();
        self
    }

    /// Add a tuple element.
    pub fn tup_element(mut self, elem: TupElement) -> Self {
        self.tuple.push(elem);
        self
    }

    /// Add an integer tuple element.
    pub fn tup_int(mut self, i: i64) -> Self {
        self.tuple.push(TupElement::Int(i));
        self
    }

    /// Add a string tuple element.
    pub fn tup_str<S: Into<String>>(mut self, s: S) -> Self {
        self.tuple.push(TupElement::String(s.into()));
        self
    }

    /// Set the value.
    pub fn value(mut self, val: Value) -> Self {
        self.value = Some(val);
        self
    }

    /// Build the KeyValue.
    ///
    /// # Panics
    ///
    /// Panics if value is not set.
    pub fn build(self) -> KeyValue {
        KeyValue {
            key: Key {
                directory: self.directory,
                tuple: self.tuple,
            },
            value: self.value.expect("value must be set before building"),
        }
    }

    /// Build the KeyValue, returning None if value is not set.
    pub fn try_build(self) -> Option<KeyValue> {
        Some(KeyValue {
            key: Key {
                directory: self.directory,
                tuple: self.tuple,
            },
            value: self.value?,
        })
    }
}

/// Builder for constructing Query instances.
#[derive(Debug, Clone, Default)]
pub struct QueryBuilder {
    directory: Directory,
    tuple: Tuple,
    value: Option<Value>,
}

impl QueryBuilder {
    /// Create a new QueryBuilder.
    pub fn new() -> Self {
        Self::default()
    }

    /// Set the directory path.
    pub fn directory<D: Into<Directory>>(mut self, dir: D) -> Self {
        self.directory = dir.into();
        self
    }

    /// Add a directory element.
    pub fn dir_element(mut self, elem: DirElement) -> Self {
        self.directory.push(elem);
        self
    }

    /// Add a string directory element.
    pub fn dir_str<S: Into<String>>(mut self, s: S) -> Self {
        self.directory.push(DirElement::String(s.into()));
        self
    }

    /// Set the tuple.
    pub fn tuple<T: Into<Tuple>>(mut self, tup: T) -> Self {
        self.tuple = tup.into();
        self
    }

    /// Add a tuple element.
    pub fn tup_element(mut self, elem: TupElement) -> Self {
        self.tuple.push(elem);
        self
    }

    /// Add an integer tuple element.
    pub fn tup_int(mut self, i: i64) -> Self {
        self.tuple.push(TupElement::Int(i));
        self
    }

    /// Add a string tuple element.
    pub fn tup_str<S: Into<String>>(mut self, s: S) -> Self {
        self.tuple.push(TupElement::String(s.into()));
        self
    }

    /// Set the value for a KeyValue query.
    pub fn value(mut self, val: Value) -> Self {
        self.value = Some(val);
        self
    }

    /// Build a KeyValue query.
    ///
    /// # Panics
    ///
    /// Panics if value is not set.
    pub fn build_keyvalue(self) -> Query {
        Query::KeyValue(KeyValue {
            key: Key {
                directory: self.directory,
                tuple: self.tuple,
            },
            value: self.value.expect("value must be set for KeyValue query"),
        })
    }

    /// Build a Key query.
    pub fn build_key(self) -> Query {
        Query::Key(Key {
            directory: self.directory,
            tuple: self.tuple,
        })
    }

    /// Build a Directory query.
    pub fn build_directory(self) -> Query {
        Query::Directory(self.directory)
    }
}

impl KeyValue {
    /// Create a builder for KeyValue.
    pub fn builder() -> KeyValueBuilder {
        KeyValueBuilder::new()
    }
}

impl Query {
    /// Create a builder for Query.
    pub fn builder() -> QueryBuilder {
        QueryBuilder::new()
    }
}

// Convenience From implementations for ergonomic construction

impl From<&str> for DirElement {
    fn from(s: &str) -> Self {
        DirElement::String(s.to_string())
    }
}

impl From<String> for DirElement {
    fn from(s: String) -> Self {
        DirElement::String(s)
    }
}

impl From<Variable> for DirElement {
    fn from(var: Variable) -> Self {
        DirElement::Variable(var)
    }
}

impl From<i64> for TupElement {
    fn from(i: i64) -> Self {
        TupElement::Int(i)
    }
}

impl From<u64> for TupElement {
    fn from(u: u64) -> Self {
        TupElement::Uint(u)
    }
}

impl From<bool> for TupElement {
    fn from(b: bool) -> Self {
        TupElement::Bool(b)
    }
}

impl From<f64> for TupElement {
    fn from(f: f64) -> Self {
        TupElement::Float(f)
    }
}

impl From<String> for TupElement {
    fn from(s: String) -> Self {
        TupElement::String(s)
    }
}

impl From<&str> for TupElement {
    fn from(s: &str) -> Self {
        TupElement::String(s.to_string())
    }
}

impl From<uuid::Uuid> for TupElement {
    fn from(u: uuid::Uuid) -> Self {
        TupElement::Uuid(u)
    }
}

impl From<Vec<u8>> for TupElement {
    fn from(b: Vec<u8>) -> Self {
        TupElement::Bytes(b)
    }
}

impl From<Tuple> for TupElement {
    fn from(t: Tuple) -> Self {
        TupElement::Tuple(t)
    }
}

impl From<Variable> for TupElement {
    fn from(var: Variable) -> Self {
        TupElement::Variable(var)
    }
}

impl From<i64> for Value {
    fn from(i: i64) -> Self {
        Value::Int(i)
    }
}

impl From<u64> for Value {
    fn from(u: u64) -> Self {
        Value::Uint(u)
    }
}

impl From<bool> for Value {
    fn from(b: bool) -> Self {
        Value::Bool(b)
    }
}

impl From<f64> for Value {
    fn from(f: f64) -> Self {
        Value::Float(f)
    }
}

impl From<String> for Value {
    fn from(s: String) -> Self {
        Value::String(s)
    }
}

impl From<&str> for Value {
    fn from(s: &str) -> Self {
        Value::String(s.to_string())
    }
}

impl From<uuid::Uuid> for Value {
    fn from(u: uuid::Uuid) -> Self {
        Value::Uuid(u)
    }
}

impl From<Vec<u8>> for Value {
    fn from(b: Vec<u8>) -> Self {
        Value::Bytes(b)
    }
}

impl From<Tuple> for Value {
    fn from(t: Tuple) -> Self {
        Value::Tuple(t)
    }
}

impl From<Variable> for Value {
    fn from(var: Variable) -> Self {
        Value::Variable(var)
    }
}

// Convenience constructors using Into

impl Key {
    /// Create a new Key with flexible input types.
    pub fn from_parts<D, T>(directory: D, tuple: T) -> Self
    where
        D: IntoIterator,
        D::Item: Into<DirElement>,
        T: IntoIterator,
        T::Item: Into<TupElement>,
    {
        Self {
            directory: directory.into_iter().map(Into::into).collect(),
            tuple: tuple.into_iter().map(Into::into).collect(),
        }
    }
}

impl KeyValue {
    /// Create a new KeyValue with flexible input types.
    pub fn from_parts<D, T, V>(directory: D, tuple: T, value: V) -> Self
    where
        D: IntoIterator,
        D::Item: Into<DirElement>,
        T: IntoIterator,
        T::Item: Into<TupElement>,
        V: Into<Value>,
    {
        Self {
            key: Key::from_parts(directory, tuple),
            value: value.into(),
        }
    }
}

// TryFrom conversions for type transformations

/// Error type for TryFrom conversions.
#[derive(Debug, Clone, PartialEq)]
pub struct ConversionError {
    pub message: String,
}

impl fmt::Display for ConversionError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Conversion error: {}", self.message)
    }
}

impl std::error::Error for ConversionError {}

// Validation functions

/// Error type for validation failures.
#[derive(Debug, Clone, PartialEq)]
pub enum ValidationError {
    /// MaybeMore must be the last element in a tuple.
    MaybeMoreNotLast,
    /// Variables are not allowed in constant values.
    VariableInConstant,
    /// Clear value is only valid for write operations.
    InvalidClearUsage,
    /// Empty tuple is invalid in this context.
    EmptyTuple,
    /// Custom validation error.
    Custom(String),
}

impl fmt::Display for ValidationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ValidationError::MaybeMoreNotLast => {
                write!(f, "MaybeMore (...) must be the last element in a tuple")
            }
            ValidationError::VariableInConstant => {
                write!(f, "Variables are not allowed in constant values")
            }
            ValidationError::InvalidClearUsage => {
                write!(f, "Clear value is only valid for write operations")
            }
            ValidationError::EmptyTuple => {
                write!(f, "Empty tuple is not allowed in this context")
            }
            ValidationError::Custom(msg) => write!(f, "Validation error: {}", msg),
        }
    }
}

impl std::error::Error for ValidationError {}

/// Validate that a tuple is well-formed.
pub fn validate_tuple(tuple: &Tuple) -> Result<(), ValidationError> {
    // Check for MaybeMore only at the end
    for (i, elem) in tuple.iter().enumerate() {
        if matches!(elem, TupElement::MaybeMore) && i != tuple.len() - 1 {
            return Err(ValidationError::MaybeMoreNotLast);
        }
    }
    Ok(())
}

/// Check if a tuple contains any variables.
pub fn tuple_has_variables(tuple: &Tuple) -> bool {
    tuple.iter().any(|elem| elem.is_variable())
}

impl TupElement {
    /// Check if this element is a variable.
    pub fn is_variable(&self) -> bool {
        match self {
            TupElement::Variable(_) => true,
            TupElement::Tuple(t) => tuple_has_variables(t),
            _ => false,
        }
    }
}

impl Value {
    /// Check if this value is a variable.
    pub fn is_variable(&self) -> bool {
        match self {
            Value::Variable(_) => true,
            Value::Tuple(t) => tuple_has_variables(t),
            _ => false,
        }
    }

    /// Check if this value is Clear.
    pub fn is_clear(&self) -> bool {
        matches!(self, Value::Clear)
    }
}

impl KeyValue {
    /// Validate that a KeyValue is well-formed.
    pub fn validate(&self) -> Result<(), ValidationError> {
        // Validate tuple structure
        validate_tuple(&self.key.tuple)?;
        Ok(())
    }

    /// Check if this is a write operation (has no variables).
    pub fn is_write(&self) -> bool {
        !tuple_has_variables(&self.key.tuple) && !self.value.is_variable()
    }

    /// Check if this is a read operation (has variables).
    pub fn is_read(&self) -> bool {
        tuple_has_variables(&self.key.tuple) || self.value.is_variable()
    }
}

impl Query {
    /// Validate that a query is well-formed.
    pub fn validate(&self) -> Result<(), ValidationError> {
        match self {
            Query::KeyValue(kv) => kv.validate(),
            Query::Key(key) => validate_tuple(&key.tuple),
            Query::Directory(_) => Ok(()),
        }
    }
}

impl TryFrom<Value> for TupElement {
    type Error = ConversionError;

    fn try_from(value: Value) -> Result<Self, Self::Error> {
        match value {
            Value::Tuple(t) => Ok(TupElement::Tuple(t)),
            Value::Nil => Ok(TupElement::Nil),
            Value::Int(i) => Ok(TupElement::Int(i)),
            Value::Uint(u) => Ok(TupElement::Uint(u)),
            Value::Bool(b) => Ok(TupElement::Bool(b)),
            Value::Float(f) => Ok(TupElement::Float(f)),
            Value::String(s) => Ok(TupElement::String(s)),
            Value::Uuid(u) => Ok(TupElement::Uuid(u)),
            Value::Bytes(b) => Ok(TupElement::Bytes(b)),
            Value::Variable(v) => Ok(TupElement::Variable(v)),
            Value::VStamp(v) => Ok(TupElement::VStamp(v)),
            Value::VStampFuture(v) => Ok(TupElement::VStampFuture(v)),
            Value::Clear => Err(ConversionError {
                message: "Cannot convert Clear value to TupElement".to_string(),
            }),
        }
    }
}

impl TryFrom<TupElement> for Value {
    type Error = ConversionError;

    fn try_from(elem: TupElement) -> Result<Self, Self::Error> {
        match elem {
            TupElement::Tuple(t) => Ok(Value::Tuple(t)),
            TupElement::Nil => Ok(Value::Nil),
            TupElement::Int(i) => Ok(Value::Int(i)),
            TupElement::Uint(u) => Ok(Value::Uint(u)),
            TupElement::Bool(b) => Ok(Value::Bool(b)),
            TupElement::Float(f) => Ok(Value::Float(f)),
            TupElement::String(s) => Ok(Value::String(s)),
            TupElement::Uuid(u) => Ok(Value::Uuid(u)),
            TupElement::Bytes(b) => Ok(Value::Bytes(b)),
            TupElement::Variable(v) => Ok(Value::Variable(v)),
            TupElement::VStamp(v) => Ok(Value::VStamp(v)),
            TupElement::VStampFuture(v) => Ok(Value::VStampFuture(v)),
            TupElement::MaybeMore => Err(ConversionError {
                message: "Cannot convert MaybeMore to Value".to_string(),
            }),
        }
    }
}

/// Type constraints for variables.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ValueType {
    /// Any type is allowed (default)
    #[serde(rename = "")]
    Any,
    /// Signed integer
    Int,
    /// Unsigned integer
    Uint,
    /// Boolean
    Bool,
    /// Floating point
    Float,
    /// String
    String,
    /// Bytes
    Bytes,
    /// UUID
    Uuid,
    /// Tuple
    Tuple,
    /// Versionstamp
    VStamp,
}

impl ValueType {
    /// Get all possible value types.
    pub fn all() -> Vec<ValueType> {
        vec![
            ValueType::Any,
            ValueType::Int,
            ValueType::Uint,
            ValueType::Bool,
            ValueType::Float,
            ValueType::String,
            ValueType::Bytes,
            ValueType::Uuid,
            ValueType::Tuple,
            ValueType::VStamp,
        ]
    }
}

impl fmt::Display for ValueType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ValueType::Any => write!(f, ""),
            ValueType::Int => write!(f, "int"),
            ValueType::Uint => write!(f, "uint"),
            ValueType::Bool => write!(f, "bool"),
            ValueType::Float => write!(f, "float"),
            ValueType::String => write!(f, "string"),
            ValueType::Bytes => write!(f, "bytes"),
            ValueType::Uuid => write!(f, "uuid"),
            ValueType::Tuple => write!(f, "tuple"),
            ValueType::VStamp => write!(f, "vstamp"),
        }
    }
}

/// A versionstamp represents a point in time in the database.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub struct VStamp {
    /// The 10-byte transaction version assigned by FoundationDB
    pub tx_version: [u8; 10],
    /// User-defined 16-bit version
    pub user_version: u16,
}

/// A future versionstamp that will be assigned when the transaction commits.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub struct VStampFuture {
    /// User-defined 16-bit version
    pub user_version: u16,
}

// Convenience constructors
impl Query {
    pub fn key_value(key: Key, value: Value) -> Self {
        Query::KeyValue(KeyValue { key, value })
    }

    pub fn key(directory: Directory, tuple: Tuple) -> Self {
        Query::Key(Key { directory, tuple })
    }

    pub fn directory(directory: Directory) -> Self {
        Query::Directory(directory)
    }
}

impl KeyValue {
    pub fn new(key: Key, value: Value) -> Self {
        Self { key, value }
    }
}

impl Key {
    pub fn new(directory: Directory, tuple: Tuple) -> Self {
        Self { directory, tuple }
    }
}

impl Default for Key {
    fn default() -> Self {
        Self {
            directory: Vec::new(),
            tuple: Vec::new(),
        }
    }
}

// Display implementations for pretty printing
impl fmt::Display for Query {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Query::KeyValue(kv) => write!(f, "{}", kv),
            Query::Key(key) => write!(f, "{}", key),
            Query::Directory(dir) => {
                for elem in dir {
                    write!(f, "/{}", elem)?;
                }
                Ok(())
            }
        }
    }
}

impl fmt::Display for KeyValue {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}={}", self.key, self.value)
    }
}

impl fmt::Display for Key {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // Format directory
        for elem in &self.directory {
            write!(f, "/{}", elem)?;
        }
        // Format tuple
        write!(f, "(")?;
        for (i, elem) in self.tuple.iter().enumerate() {
            if i > 0 {
                write!(f, ",")?;
            }
            write!(f, "{}", elem)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for DirElement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            DirElement::String(s) => write!(f, "{}", s),
            DirElement::Variable(var) => write!(f, "<{}>", var),
        }
    }
}

impl fmt::Display for TupElement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            TupElement::Tuple(t) => {
                write!(f, "(")?;
                for (i, elem) in t.iter().enumerate() {
                    if i > 0 {
                        write!(f, ",")?;
                    }
                    write!(f, "{}", elem)?;
                }
                write!(f, ")")
            }
            TupElement::Nil => write!(f, "nil"),
            TupElement::Int(i) => write!(f, "{}", i),
            TupElement::Uint(u) => write!(f, "{}", u),
            TupElement::Bool(b) => write!(f, "{}", b),
            TupElement::Float(fl) => write!(f, "{}", fl),
            TupElement::String(s) => write!(f, "\"{}\"", s),
            TupElement::Uuid(u) => write!(f, "{}", u),
            TupElement::Bytes(b) => {
                write!(f, "0x")?;
                for byte in b {
                    write!(f, "{:02x}", byte)?;
                }
                Ok(())
            }
            TupElement::Variable(var) => write!(f, "<{}>", var),
            TupElement::MaybeMore => write!(f, "..."),
            TupElement::VStamp(_) => write!(f, "#vstamp"),
            TupElement::VStampFuture(_) => write!(f, "#vstamp_future"),
        }
    }
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::Tuple(t) => {
                write!(f, "(")?;
                for (i, elem) in t.iter().enumerate() {
                    if i > 0 {
                        write!(f, ",")?;
                    }
                    write!(f, "{}", elem)?;
                }
                write!(f, ")")
            }
            Value::Nil => write!(f, "nil"),
            Value::Int(i) => write!(f, "{}", i),
            Value::Uint(u) => write!(f, "{}", u),
            Value::Bool(b) => write!(f, "{}", b),
            Value::Float(fl) => write!(f, "{}", fl),
            Value::String(s) => write!(f, "\"{}\"", s),
            Value::Uuid(u) => write!(f, "{}", u),
            Value::Bytes(b) => {
                write!(f, "0x")?;
                for byte in b {
                    write!(f, "{:02x}", byte)?;
                }
                Ok(())
            }
            Value::Variable(var) => write!(f, "<{}>", var),
            Value::Clear => write!(f, "clear"),
            Value::VStamp(_) => write!(f, "#vstamp"),
            Value::VStampFuture(_) => write!(f, "#vstamp_future"),
        }
    }
}

impl fmt::Display for Variable {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.types.is_empty() {
            Ok(())
        } else {
            let types: Vec<String> = self.types.iter().map(|t| t.to_string()).collect();
            write!(f, "{}", types.join("|"))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_variable_any() {
        let var = Variable::any();
        assert_eq!(var.types.len(), 0);
    }

    #[test]
    fn test_variable_with_types() {
        let var = Variable::with_types(vec![ValueType::Int, ValueType::String]);
        assert_eq!(var.types.len(), 2);
    }

    #[test]
    fn test_query_constructors() {
        let dir = vec![DirElement::String("test".to_string())];
        let tuple = vec![TupElement::Int(42)];
        let key = Key::new(dir.clone(), tuple);
        let value = Value::Int(100);

        let q1 = Query::key_value(key.clone(), value);
        match q1 {
            Query::KeyValue(kv) => {
                assert_eq!(kv.value, Value::Int(100));
            }
            _ => panic!("Expected KeyValue"),
        }

        let q2 = Query::key(dir.clone(), vec![]);
        match q2 {
            Query::Key(_) => {}
            _ => panic!("Expected Key"),
        }

        let q3 = Query::directory(dir);
        match q3 {
            Query::Directory(_) => {}
            _ => panic!("Expected Directory"),
        }
    }

    #[test]
    fn test_display_query() {
        let query = Query::KeyValue(KeyValue {
            key: Key {
                directory: vec![DirElement::String("users".to_string())],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::String("data".to_string()),
        });
        assert_eq!(query.to_string(), "/users(42)=\"data\"");
    }

    #[test]
    fn test_display_key() {
        let key = Key {
            directory: vec![
                DirElement::String("db".to_string()),
                DirElement::String("users".to_string()),
            ],
            tuple: vec![TupElement::Int(1), TupElement::String("key".to_string())],
        };
        assert_eq!(key.to_string(), "/db/users(1,\"key\")");
    }

    #[test]
    fn test_display_value() {
        assert_eq!(Value::Int(42).to_string(), "42");
        assert_eq!(Value::String("hello".to_string()).to_string(), "\"hello\"");
        assert_eq!(Value::Bool(true).to_string(), "true");
        assert_eq!(Value::Nil.to_string(), "nil");
        assert_eq!(Value::Clear.to_string(), "clear");
    }

    #[test]
    fn test_display_tuple_element() {
        assert_eq!(TupElement::Int(99).to_string(), "99");
        assert_eq!(TupElement::String("test".to_string()).to_string(), "\"test\"");
        assert_eq!(TupElement::MaybeMore.to_string(), "...");

        let nested = TupElement::Tuple(vec![TupElement::Int(1), TupElement::Int(2)]);
        assert_eq!(nested.to_string(), "(1,2)");
    }

    #[test]
    fn test_display_variable() {
        let var1 = Variable::any();
        assert_eq!(var1.to_string(), "");

        let var2 = Variable::with_types(vec![ValueType::Int, ValueType::String]);
        assert_eq!(var2.to_string(), "int|string");
    }

    #[test]
    fn test_display_bytes() {
        let bytes = TupElement::Bytes(vec![0xde, 0xad, 0xbe, 0xef]);
        assert_eq!(bytes.to_string(), "0xdeadbeef");
    }

    #[test]
    fn test_keyvalue_builder_basic() {
        let kv = KeyValue::builder()
            .dir_str("users")
            .tup_int(42)
            .value(Value::String("data".to_string()))
            .build();

        assert_eq!(kv.key.directory.len(), 1);
        assert_eq!(kv.key.tuple.len(), 1);
        assert_eq!(kv.value, Value::String("data".to_string()));
    }

    #[test]
    fn test_keyvalue_builder_chaining() {
        let kv = KeyValueBuilder::new()
            .dir_str("db")
            .dir_str("users")
            .tup_int(1)
            .tup_str("key")
            .value(Value::Bool(true))
            .build();

        assert_eq!(kv.key.directory.len(), 2);
        assert_eq!(kv.key.tuple.len(), 2);
    }

    #[test]
    fn test_keyvalue_builder_try_build() {
        let kv = KeyValue::builder()
            .dir_str("test")
            .tup_int(1)
            .value(Value::Nil)
            .try_build();

        assert!(kv.is_some());

        let empty = KeyValue::builder().dir_str("test").try_build();
        assert!(empty.is_none());
    }

    #[test]
    #[should_panic(expected = "value must be set before building")]
    fn test_keyvalue_builder_panic() {
        KeyValue::builder().dir_str("test").build();
    }

    #[test]
    fn test_query_builder_keyvalue() {
        let query = Query::builder()
            .dir_str("users")
            .tup_int(42)
            .value(Value::Int(100))
            .build_keyvalue();

        match query {
            Query::KeyValue(kv) => {
                assert_eq!(kv.key.directory.len(), 1);
                assert_eq!(kv.value, Value::Int(100));
            }
            _ => panic!("Expected KeyValue query"),
        }
    }

    #[test]
    fn test_query_builder_key() {
        let query = Query::builder()
            .dir_str("test")
            .tup_str("key")
            .build_key();

        match query {
            Query::Key(key) => {
                assert_eq!(key.directory.len(), 1);
                assert_eq!(key.tuple.len(), 1);
            }
            _ => panic!("Expected Key query"),
        }
    }

    #[test]
    fn test_query_builder_directory() {
        let query = Query::builder()
            .dir_str("users")
            .dir_str("profiles")
            .build_directory();

        match query {
            Query::Directory(dir) => {
                assert_eq!(dir.len(), 2);
            }
            _ => panic!("Expected Directory query"),
        }
    }

    #[test]
    fn test_from_conversions_tup_element() {
        assert!(matches!(TupElement::from(42i64), TupElement::Int(42)));
        assert!(matches!(TupElement::from(42u64), TupElement::Uint(42)));
        assert!(matches!(TupElement::from(true), TupElement::Bool(true)));
        assert!(matches!(TupElement::from(3.14f64), TupElement::Float(_)));
        assert!(matches!(
            TupElement::from("test"),
            TupElement::String(s) if s == "test"
        ));
    }

    #[test]
    fn test_from_conversions_value() {
        assert!(matches!(Value::from(42i64), Value::Int(42)));
        assert!(matches!(Value::from(42u64), Value::Uint(42)));
        assert!(matches!(Value::from(true), Value::Bool(true)));
        assert!(matches!(Value::from(3.14f64), Value::Float(_)));
        assert!(matches!(
            Value::from("test"),
            Value::String(s) if s == "test"
        ));
    }

    #[test]
    fn test_from_conversions_dir_element() {
        assert!(matches!(
            DirElement::from("test"),
            DirElement::String(s) if s == "test"
        ));
        let var = Variable::any();
        assert!(matches!(DirElement::from(var), DirElement::Variable(_)));
    }

    #[test]
    fn test_key_from_parts() {
        let key = Key::from_parts(
            vec!["users", "profiles"],
            vec![TupElement::Int(42), TupElement::String("key".to_string())],
        );
        assert_eq!(key.directory.len(), 2);
        assert_eq!(key.tuple.len(), 2);
        assert!(matches!(key.tuple[0], TupElement::Int(42)));
    }

    #[test]
    fn test_keyvalue_from_parts() {
        let kv = KeyValue::from_parts(
            vec!["users"],
            vec![42i64],
            "data"
        );
        assert_eq!(kv.key.directory.len(), 1);
        assert_eq!(kv.key.tuple.len(), 1);
        assert!(matches!(kv.value, Value::String(s) if s == "data"));
    }

    #[test]
    fn test_try_from_value_to_tup_element() {
        use std::convert::TryFrom;

        let val = Value::Int(42);
        let elem = TupElement::try_from(val).unwrap();
        assert!(matches!(elem, TupElement::Int(42)));

        let val = Value::String("test".to_string());
        let elem = TupElement::try_from(val).unwrap();
        assert!(matches!(elem, TupElement::String(s) if s == "test"));

        // Test error case
        let val = Value::Clear;
        let result = TupElement::try_from(val);
        assert!(result.is_err());
    }

    #[test]
    fn test_try_from_tup_element_to_value() {
        use std::convert::TryFrom;

        let elem = TupElement::Int(42);
        let val = Value::try_from(elem).unwrap();
        assert!(matches!(val, Value::Int(42)));

        let elem = TupElement::String("test".to_string());
        let val = Value::try_from(elem).unwrap();
        assert!(matches!(val, Value::String(s) if s == "test"));

        // Test error case
        let elem = TupElement::MaybeMore;
        let result = Value::try_from(elem);
        assert!(result.is_err());
    }

    #[test]
    fn test_default_implementations() {
        let var = Variable::default();
        assert_eq!(var.types.len(), 0);
        assert_eq!(var, Variable::any());

        let key = Key::default();
        assert_eq!(key.directory.len(), 0);
        assert_eq!(key.tuple.len(), 0);
    }

    #[test]
    fn test_validation_maybe_more_position() {
        // Valid: MaybeMore at the end
        let tuple = vec![TupElement::Int(1), TupElement::MaybeMore];
        assert!(validate_tuple(&tuple).is_ok());

        // Invalid: MaybeMore in the middle
        let tuple = vec![TupElement::MaybeMore, TupElement::Int(1)];
        assert!(matches!(
            validate_tuple(&tuple),
            Err(ValidationError::MaybeMoreNotLast)
        ));
    }

    #[test]
    fn test_validation_has_variables() {
        let tuple_with_var = vec![TupElement::Int(1), TupElement::Variable(Variable::any())];
        assert!(tuple_has_variables(&tuple_with_var));

        let tuple_no_var = vec![TupElement::Int(1), TupElement::String("test".to_string())];
        assert!(!tuple_has_variables(&tuple_no_var));

        let nested_var = vec![TupElement::Tuple(vec![TupElement::Variable(Variable::any())])];
        assert!(tuple_has_variables(&nested_var));
    }

    #[test]
    fn test_validation_value_checks() {
        let val = Value::Variable(Variable::any());
        assert!(val.is_variable());

        let val = Value::Clear;
        assert!(val.is_clear());

        let val = Value::Int(42);
        assert!(!val.is_variable());
        assert!(!val.is_clear());
    }

    #[test]
    fn test_validation_keyvalue_operations() {
        // Write operation: no variables
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::String("data".to_string()),
        };
        assert!(kv.is_write());
        assert!(!kv.is_read());

        // Read operation: has variable
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Variable(Variable::any())],
            },
            value: Value::String("data".to_string()),
        };
        assert!(!kv.is_write());
        assert!(kv.is_read());
    }

    #[test]
    fn test_validation_query() {
        let query = Query::KeyValue(KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(1), TupElement::MaybeMore],
            },
            value: Value::Int(42),
        });
        assert!(query.validate().is_ok());

        let query = Query::KeyValue(KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::MaybeMore, TupElement::Int(1)],
            },
            value: Value::Int(42),
        });
        assert!(query.validate().is_err());
    }
}
