//! Value serialization and deserialization.
//!
//! This module provides serialization and deserialization of FQL values to/from bytes
//! for storage in FoundationDB.

use crate::*;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum SerializationError {
    #[error("Serialization failed: {0}")]
    Failed(String),

    #[error("Cannot serialize {0}")]
    CannotSerialize(String),

    #[error("Invalid data length: expected {expected}, got {actual}")]
    InvalidLength { expected: usize, actual: usize },

    #[error("Unknown value type: {0}")]
    UnknownType(String),
}

/// Endianness configuration for numeric value serialization.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Endianness {
    Big,
    Little,
}

/// Pack a value into bytes for storage.
///
/// # Arguments
/// * `value` - The value to serialize
/// * `endianness` - Byte order for numeric types
/// * `has_vstamp` - Whether the value contains a versionstamp
///
/// # Returns
/// The serialized bytes, or an error if serialization fails.
///
/// # Errors
/// Returns an error if:
/// - The value is a Variable or Clear (cannot be serialized)
/// - Tuple conversion fails
pub fn pack(value: &Value, endianness: Endianness, has_vstamp: bool) -> Result<Vec<u8>, SerializationError> {
    match value {
        Value::Nil => Ok(vec![]),

        Value::Bool(b) => Ok(vec![if *b { 1 } else { 0 }]),

        Value::Int(i) => {
            let mut bytes = vec![0u8; 8];
            match endianness {
                Endianness::Big => {
                    bytes.copy_from_slice(&(*i as u64).to_be_bytes());
                }
                Endianness::Little => {
                    bytes.copy_from_slice(&(*i as u64).to_le_bytes());
                }
            }
            Ok(bytes)
        }

        Value::Uint(u) => {
            let mut bytes = vec![0u8; 8];
            match endianness {
                Endianness::Big => {
                    bytes.copy_from_slice(&u.to_be_bytes());
                }
                Endianness::Little => {
                    bytes.copy_from_slice(&u.to_le_bytes());
                }
            }
            Ok(bytes)
        }

        Value::Float(f) => {
            let bits = f.to_bits();
            let mut bytes = vec![0u8; 8];
            match endianness {
                Endianness::Big => {
                    bytes.copy_from_slice(&bits.to_be_bytes());
                }
                Endianness::Little => {
                    bytes.copy_from_slice(&bits.to_le_bytes());
                }
            }
            Ok(bytes)
        }

        Value::String(s) => Ok(s.as_bytes().to_vec()),

        Value::Bytes(b) => Ok(b.clone()),

        Value::Uuid(u) => Ok(u.as_bytes().to_vec()),

        Value::Tuple(_tup) => {
            // TODO: Implement FDB tuple packing
            // For now, use a simple placeholder serialization
            // In a real implementation, this would use FoundationDB's tuple layer
            if has_vstamp {
                Err(SerializationError::Failed(
                    "Versionstamp tuple packing not yet implemented".to_string()
                ))
            } else {
                // Placeholder: just serialize as empty for now
                Ok(vec![])
            }
        }

        Value::VStamp(vstamp) => {
            // VStamp is 12 bytes: 10 bytes tx_version + 2 bytes user_version
            let mut bytes = vec![0u8; 12];
            bytes[0..10].copy_from_slice(&vstamp.tx_version);
            bytes[10..12].copy_from_slice(&vstamp.user_version.to_le_bytes());
            Ok(bytes)
        }

        Value::VStampFuture(vstamp) => {
            // VStampFuture is 16 bytes: 10 bytes (zeros) + 2 bytes user_version + 4 bytes position
            let mut bytes = vec![0u8; 16];
            bytes[10..12].copy_from_slice(&vstamp.user_version.to_le_bytes());
            Ok(bytes)
        }

        Value::Variable(_) => {
            Err(SerializationError::CannotSerialize("Variable".to_string()))
        }

        Value::Clear => {
            Err(SerializationError::CannotSerialize("Clear".to_string()))
        }
    }
}

/// Unpack bytes into a value with a type hint.
///
/// # Arguments
/// * `data` - The bytes to deserialize
/// * `value_type` - Type hint for deserialization
/// * `endianness` - Byte order for numeric types
///
/// # Returns
/// The deserialized value, or an error if deserialization fails.
///
/// # Errors
/// Returns an error if:
/// - The data length doesn't match the expected length for the type
/// - The value type is unknown
pub fn unpack(data: &[u8], value_type: ValueType, endianness: Endianness) -> Result<Value, SerializationError> {
    match value_type {
        ValueType::Any => Ok(Value::Bytes(data.to_vec())),

        ValueType::Bool => {
            if data.len() != 1 {
                return Err(SerializationError::InvalidLength {
                    expected: 1,
                    actual: data.len(),
                });
            }
            Ok(Value::Bool(data[0] == 1))
        }

        ValueType::Int => {
            if data.len() != 8 {
                return Err(SerializationError::InvalidLength {
                    expected: 8,
                    actual: data.len(),
                });
            }
            let value = match endianness {
                Endianness::Big => u64::from_be_bytes(data.try_into().unwrap()),
                Endianness::Little => u64::from_le_bytes(data.try_into().unwrap()),
            };
            Ok(Value::Int(value as i64))
        }

        ValueType::Uint => {
            if data.len() != 8 {
                return Err(SerializationError::InvalidLength {
                    expected: 8,
                    actual: data.len(),
                });
            }
            let value = match endianness {
                Endianness::Big => u64::from_be_bytes(data.try_into().unwrap()),
                Endianness::Little => u64::from_le_bytes(data.try_into().unwrap()),
            };
            Ok(Value::Uint(value))
        }

        ValueType::Float => {
            if data.len() != 8 {
                return Err(SerializationError::InvalidLength {
                    expected: 8,
                    actual: data.len(),
                });
            }
            let bits = match endianness {
                Endianness::Big => u64::from_be_bytes(data.try_into().unwrap()),
                Endianness::Little => u64::from_le_bytes(data.try_into().unwrap()),
            };
            Ok(Value::Float(f64::from_bits(bits)))
        }

        ValueType::String => {
            let s = String::from_utf8(data.to_vec())
                .map_err(|e| SerializationError::Failed(format!("Invalid UTF-8: {}", e)))?;
            Ok(Value::String(s))
        }

        ValueType::Bytes => Ok(Value::Bytes(data.to_vec())),

        ValueType::Uuid => {
            if data.len() != 16 {
                return Err(SerializationError::InvalidLength {
                    expected: 16,
                    actual: data.len(),
                });
            }
            let uuid = uuid::Uuid::from_slice(data)
                .map_err(|e| SerializationError::Failed(format!("Invalid UUID: {}", e)))?;
            Ok(Value::Uuid(uuid))
        }

        ValueType::Tuple => {
            // TODO: Implement FDB tuple unpacking
            // For now, return an empty tuple as placeholder
            Ok(Value::Tuple(vec![]))
        }

        ValueType::VStamp => {
            if data.len() != 12 {
                return Err(SerializationError::InvalidLength {
                    expected: 12,
                    actual: data.len(),
                });
            }
            let mut tx_version = [0u8; 10];
            tx_version.copy_from_slice(&data[0..10]);
            let user_version = u16::from_le_bytes([data[10], data[11]]);
            Ok(Value::VStamp(VStamp {
                tx_version,
                user_version,
            }))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_pack_unpack_nil() {
        let value = Value::Nil;
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed, vec![]);
    }

    #[test]
    fn test_pack_unpack_bool() {
        let value_true = Value::Bool(true);
        let packed_true = pack(&value_true, Endianness::Big, false).unwrap();
        assert_eq!(packed_true, vec![1]);
        let unpacked_true = unpack(&packed_true, ValueType::Bool, Endianness::Big).unwrap();
        assert_eq!(unpacked_true, value_true);

        let value_false = Value::Bool(false);
        let packed_false = pack(&value_false, Endianness::Big, false).unwrap();
        assert_eq!(packed_false, vec![0]);
        let unpacked_false = unpack(&packed_false, ValueType::Bool, Endianness::Big).unwrap();
        assert_eq!(unpacked_false, value_false);
    }

    #[test]
    fn test_pack_unpack_int() {
        let value = Value::Int(42);
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed.len(), 8);
        let unpacked = unpack(&packed, ValueType::Int, Endianness::Big).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_unpack_uint() {
        let value = Value::Uint(12345);
        let packed = pack(&value, Endianness::Little, false).unwrap();
        assert_eq!(packed.len(), 8);
        let unpacked = unpack(&packed, ValueType::Uint, Endianness::Little).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_unpack_float() {
        let value = Value::Float(3.14159);
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed.len(), 8);
        let unpacked = unpack(&packed, ValueType::Float, Endianness::Big).unwrap();
        if let Value::Float(f) = unpacked {
            assert!((f - 3.14159).abs() < 1e-10);
        } else {
            panic!("Expected Float");
        }
    }

    #[test]
    fn test_pack_unpack_string() {
        let value = Value::String("hello world".to_string());
        let packed = pack(&value, Endianness::Big, false).unwrap();
        let unpacked = unpack(&packed, ValueType::String, Endianness::Big).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_unpack_bytes() {
        let value = Value::Bytes(vec![1, 2, 3, 4, 5]);
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed, vec![1, 2, 3, 4, 5]);
        let unpacked = unpack(&packed, ValueType::Bytes, Endianness::Big).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_unpack_uuid() {
        let uuid = uuid::Uuid::new_v4();
        let value = Value::Uuid(uuid);
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed.len(), 16);
        let unpacked = unpack(&packed, ValueType::Uuid, Endianness::Big).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_unpack_vstamp() {
        let vstamp = VStamp {
            tx_version: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
            user_version: 42,
        };
        let value = Value::VStamp(vstamp);
        let packed = pack(&value, Endianness::Big, false).unwrap();
        assert_eq!(packed.len(), 12);
        let unpacked = unpack(&packed, ValueType::VStamp, Endianness::Big).unwrap();
        assert_eq!(unpacked, value);
    }

    #[test]
    fn test_pack_variable_fails() {
        let value = Value::Variable(Variable::any());
        let result = pack(&value, Endianness::Big, false);
        assert!(result.is_err());
    }

    #[test]
    fn test_pack_clear_fails() {
        let value = Value::Clear;
        let result = pack(&value, Endianness::Big, false);
        assert!(result.is_err());
    }

    #[test]
    fn test_endianness_difference() {
        let value = Value::Int(0x0102030405060708u64 as i64);

        let packed_big = pack(&value, Endianness::Big, false).unwrap();
        let packed_little = pack(&value, Endianness::Little, false).unwrap();

        // Different endianness should produce different bytes
        assert_ne!(packed_big, packed_little);

        // But unpacking with correct endianness should give same value
        let unpacked_big = unpack(&packed_big, ValueType::Int, Endianness::Big).unwrap();
        let unpacked_little = unpack(&packed_little, ValueType::Int, Endianness::Little).unwrap();

        assert_eq!(unpacked_big, value);
        assert_eq!(unpacked_little, value);
    }
}
