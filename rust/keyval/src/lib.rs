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
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Variable {
    /// Type constraints for this variable. Empty means any type is allowed.
    pub types: Vec<ValueType>,
}

impl Variable {
    /// Create a new variable that accepts any type.
    pub fn any() -> Self {
        Self { types: Vec::new() }
    }

    /// Create a new variable with specific type constraints.
    pub fn with_types(types: Vec<ValueType>) -> Self {
        Self { types }
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
}
