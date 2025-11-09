//! Query classification module.
//!
//! This module classifies a KeyValue by the kind of operation it represents.
//! Classifications include: Constant, VStampKey, VStampVal, Clear, ReadSingle, and ReadRange.

use crate::*;
use std::fmt;

/// Classification of a KeyValue based on its structure and contents.
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Class {
    /// KeyValue with no Variable, MaybeMore, Clear, or VStampFuture.
    /// Can be used for set operations or returned by get operations.
    Constant,

    /// Constant KeyValue with a VStampFuture in the key.
    /// Can only be used for set operations.
    VStampKey,

    /// Constant KeyValue with a VStampFuture in the value.
    /// Can only be used for set operations.
    VStampVal,

    /// KeyValue with Clear as its value and no Variable, MaybeMore, or VStampFuture.
    /// Used for clear/delete operations.
    Clear,

    /// KeyValue with Variable as its value but no Variable or MaybeMore in its key.
    /// Returns a single KeyValue.
    ReadSingle,

    /// KeyValue with Variable or MaybeMore in its key.
    /// Returns multiple KeyValues.
    ReadRange,

    /// Invalid KeyValue with conflicting attributes.
    Invalid(String),
}

impl fmt::Display for Class {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Class::Constant => write!(f, "constant"),
            Class::VStampKey => write!(f, "vstampkey"),
            Class::VStampVal => write!(f, "vstampval"),
            Class::Clear => write!(f, "clear"),
            Class::ReadSingle => write!(f, "single"),
            Class::ReadRange => write!(f, "range"),
            Class::Invalid(msg) => write!(f, "invalid[{}]", msg),
        }
    }
}

/// Classify a KeyValue based on its structure.
pub fn classify(kv: &KeyValue) -> Class {
    let dir_attr = get_attributes_of_dir(&kv.key.directory);
    let key_attr = dir_attr.merge(&get_attributes_of_tup(&kv.key.tuple));
    let kv_attr = key_attr.merge(&get_attributes_of_val(&kv.value));

    // KeyValues should never contain nil (in Go sense - we don't have this issue in Rust)
    // But we track it for compatibility
    if kv_attr.has_nil {
        return invalid_class(&kv_attr);
    }

    // KeyValues should contain at most 1 VStampFuture
    if kv_attr.vstamp_futures > 1 {
        return invalid_class(&kv_attr);
    }

    // Ensure at most one of these conditions is true
    let count = [
        kv_attr.vstamp_futures > 0,
        kv_attr.has_variable,
        kv_attr.has_clear,
    ]
    .iter()
    .filter(|&&x| x)
    .count();

    if count > 1 {
        return invalid_class(&kv_attr);
    }

    match () {
        _ if key_attr.has_variable => Class::ReadRange,
        _ if kv_attr.has_variable => Class::ReadSingle,
        _ if kv_attr.vstamp_futures > 0 => {
            if key_attr.vstamp_futures > 0 {
                Class::VStampKey
            } else {
                Class::VStampVal
            }
        }
        _ if kv_attr.has_clear => Class::Clear,
        _ => Class::Constant,
    }
}

/// Attributes describing the characteristics of a KeyValue relevant to classification.
#[derive(Debug, Clone, Copy, Default)]
struct Attributes {
    vstamp_futures: usize,
    has_variable: bool,
    has_clear: bool,
    has_nil: bool,
}

impl Attributes {
    /// Merge attributes of parts to infer attributes of the whole.
    fn merge(&self, other: &Attributes) -> Attributes {
        Attributes {
            vstamp_futures: self.vstamp_futures + other.vstamp_futures,
            has_variable: self.has_variable || other.has_variable,
            has_clear: self.has_clear || other.has_clear,
            has_nil: self.has_nil || other.has_nil,
        }
    }
}

/// Get attributes of a directory.
fn get_attributes_of_dir(dir: &Directory) -> Attributes {
    let mut attr = Attributes::default();
    for element in dir {
        match element {
            DirElement::String(_) => {}
            DirElement::Variable(_) => attr.has_variable = true,
        }
    }
    attr
}

/// Get attributes of a tuple.
fn get_attributes_of_tup(tup: &Tuple) -> Attributes {
    let mut attr = Attributes::default();
    for element in tup {
        let sub_attr = match element {
            TupElement::Tuple(t) => get_attributes_of_tup(t),
            TupElement::Variable(_) => Attributes {
                has_variable: true,
                ..Default::default()
            },
            TupElement::MaybeMore => Attributes {
                has_variable: true,
                ..Default::default()
            },
            TupElement::VStampFuture(_) => Attributes {
                vstamp_futures: 1,
                ..Default::default()
            },
            _ => Attributes::default(),
        };
        attr = attr.merge(&sub_attr);
    }
    attr
}

/// Get attributes of a value.
fn get_attributes_of_val(val: &Value) -> Attributes {
    match val {
        Value::Tuple(t) => get_attributes_of_tup(t),
        Value::Variable(_) => Attributes {
            has_variable: true,
            ..Default::default()
        },
        Value::Clear => Attributes {
            has_clear: true,
            ..Default::default()
        },
        Value::VStampFuture(_) => Attributes {
            vstamp_futures: 1,
            ..Default::default()
        },
        _ => Attributes::default(),
    }
}

/// Create an Invalid class with relevant attributes for debugging.
fn invalid_class(attr: &Attributes) -> Class {
    let mut parts = Vec::new();

    if attr.vstamp_futures > 0 {
        parts.push(format!("vstamps:{}", attr.vstamp_futures));
    }
    if attr.has_variable {
        parts.push("var".to_string());
    }
    if attr.has_clear {
        parts.push("clear".to_string());
    }
    if attr.has_nil {
        parts.push("nil".to_string());
    }

    Class::Invalid(parts.join(","))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_classify_constant() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::String("test".to_string()),
        };
        assert_eq!(classify(&kv), Class::Constant);
    }

    #[test]
    fn test_classify_clear() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::Clear,
        };
        assert_eq!(classify(&kv), Class::Clear);
    }

    #[test]
    fn test_classify_read_single() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::Variable(Variable::any()),
        };
        assert_eq!(classify(&kv), Class::ReadSingle);
    }

    #[test]
    fn test_classify_read_range() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42), TupElement::MaybeMore],
            },
            value: Value::Variable(Variable::any()),
        };
        assert_eq!(classify(&kv), Class::ReadRange);
    }

    #[test]
    fn test_classify_vstamp_key() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::VStampFuture(VStampFuture { user_version: 0 })],
            },
            value: Value::Int(42),
        };
        assert_eq!(classify(&kv), Class::VStampKey);
    }

    #[test]
    fn test_classify_vstamp_val() {
        let kv = KeyValue {
            key: Key {
                directory: vec![],
                tuple: vec![TupElement::Int(42)],
            },
            value: Value::VStampFuture(VStampFuture { user_version: 0 }),
        };
        assert_eq!(classify(&kv), Class::VStampVal);
    }
}
