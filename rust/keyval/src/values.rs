//! Value serialization and deserialization.

use crate::*;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum SerializationError {
    #[error("Serialization failed: {0}")]
    Failed(String),
}

/// Endianness configuration for numeric value serialization.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Endianness {
    Big,
    Little,
}

/// Pack a value into bytes.
pub fn pack(value: &Value, endianness: Endianness) -> Result<Vec<u8>, SerializationError> {
    // TODO: Implement value packing
    Ok(vec![])
}

/// Unpack bytes into a value.
pub fn unpack(data: &[u8], endianness: Endianness) -> Result<Value, SerializationError> {
    // TODO: Implement value unpacking
    Ok(Value::Nil)
}
