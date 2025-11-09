//! Tuple comparison and matching logic.
//!
//! This module provides functionality to compare candidate tuples against
//! schema tuples that may contain Variables and MaybeMore wildcards.

use crate::*;

/// Compare a candidate tuple against a schema tuple.
///
/// The schema tuple may contain Variable or MaybeMore elements, while the candidate
/// must not contain either. Returns None if the tuples match, or Some with the
/// index path to the first mismatching element.
///
/// # Examples
///
/// ```
/// use keyval::*;
/// use keyval::tuple::compare;
///
/// let schema = vec![TupElement::Int(1), TupElement::Variable(Variable::any())];
/// let candidate = vec![TupElement::Int(1), TupElement::String("hello".to_string())];
/// assert_eq!(compare(&schema, &candidate), None); // Matches
///
/// let bad_candidate = vec![TupElement::Int(2), TupElement::String("hello".to_string())];
/// assert_eq!(compare(&schema, &bad_candidate), Some(vec![0])); // First element mismatch
/// ```
pub fn compare(schema: &Tuple, candidate: &Tuple) -> Option<Vec<usize>> {
    // If the schema is empty, the candidate must be empty as well
    if schema.is_empty() {
        return if candidate.is_empty() {
            None
        } else {
            Some(vec![0])
        };
    }

    // Check if schema ends with MaybeMore
    let has_maybe_more = matches!(schema.last(), Some(TupElement::MaybeMore));
    let schema_to_check = if has_maybe_more {
        &schema[..schema.len() - 1]
    } else {
        schema
    };

    // If no MaybeMore, lengths must match or candidate must be shorter
    if !has_maybe_more && schema_to_check.len() < candidate.len() {
        return Some(vec![schema_to_check.len()]);
    }

    // Candidate must be at least as long as schema (minus MaybeMore)
    if schema_to_check.len() > candidate.len() {
        return Some(vec![candidate.len()]);
    }

    // Compare each element
    for (i, schema_elem) in schema_to_check.iter().enumerate() {
        if let Some(mismatch) = compare_element(schema_elem, &candidate[i], i) {
            return Some(mismatch);
        }
    }

    None
}

/// Compare a single tuple element against a candidate element
fn compare_element(schema: &TupElement, candidate: &TupElement, index: usize) -> Option<Vec<usize>> {
    match schema {
        TupElement::Tuple(schema_tup) => {
            if let TupElement::Tuple(cand_tup) = candidate {
                if let Some(mut mismatch) = compare(schema_tup, cand_tup) {
                    mismatch.insert(0, index);
                    return Some(mismatch);
                }
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Variable(var) => {
            // Empty variable means any type is allowed
            if var.types.is_empty() {
                return None;
            }

            // Check if candidate matches any of the allowed types
            for vtype in &var.types {
                if matches_type(candidate, *vtype) {
                    return None;
                }
            }
            Some(vec![index])
        }
        TupElement::MaybeMore => {
            // MaybeMore should have been removed before comparison
            Some(vec![index])
        }
        // For all other types, check equality
        TupElement::Nil => {
            if matches!(candidate, TupElement::Nil) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Int(v) => {
            if matches!(candidate, TupElement::Int(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Uint(v) => {
            if matches!(candidate, TupElement::Uint(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Bool(v) => {
            if matches!(candidate, TupElement::Bool(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Float(v) => {
            if matches!(candidate, TupElement::Float(c) if (c - v).abs() < f64::EPSILON) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::String(v) => {
            if matches!(candidate, TupElement::String(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Uuid(v) => {
            if matches!(candidate, TupElement::Uuid(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::Bytes(v) => {
            if matches!(candidate, TupElement::Bytes(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::VStamp(v) => {
            if matches!(candidate, TupElement::VStamp(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
        TupElement::VStampFuture(v) => {
            if matches!(candidate, TupElement::VStampFuture(c) if c == v) {
                None
            } else {
                Some(vec![index])
            }
        }
    }
}

/// Check if a tuple element matches a specific value type
fn matches_type(elem: &TupElement, vtype: ValueType) -> bool {
    match vtype {
        ValueType::Any => true,
        ValueType::Int => matches!(elem, TupElement::Int(_)),
        ValueType::Uint => matches!(elem, TupElement::Uint(_)),
        ValueType::Bool => matches!(elem, TupElement::Bool(_)),
        ValueType::Float => matches!(elem, TupElement::Float(_)),
        ValueType::String => matches!(elem, TupElement::String(_)),
        ValueType::Bytes => matches!(elem, TupElement::Bytes(_)),
        ValueType::Uuid => matches!(elem, TupElement::Uuid(_)),
        ValueType::Tuple => matches!(elem, TupElement::Tuple(_)),
        ValueType::VStamp => matches!(elem, TupElement::VStamp(_)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compare_empty_tuples() {
        let schema = vec![];
        let candidate = vec![];
        assert_eq!(compare(&schema, &candidate), None);
    }

    #[test]
    fn test_compare_empty_schema_nonempty_candidate() {
        let schema = vec![];
        let candidate = vec![TupElement::Int(1)];
        assert_eq!(compare(&schema, &candidate), Some(vec![0]));
    }

    #[test]
    fn test_compare_exact_match() {
        let schema = vec![TupElement::Int(42), TupElement::String("hello".to_string())];
        let candidate = vec![TupElement::Int(42), TupElement::String("hello".to_string())];
        assert_eq!(compare(&schema, &candidate), None);
    }

    #[test]
    fn test_compare_mismatch() {
        let schema = vec![TupElement::Int(42), TupElement::String("hello".to_string())];
        let candidate = vec![TupElement::Int(43), TupElement::String("hello".to_string())];
        assert_eq!(compare(&schema, &candidate), Some(vec![0]));
    }

    #[test]
    fn test_compare_with_variable_any() {
        let schema = vec![TupElement::Int(42), TupElement::Variable(Variable::any())];
        let candidate = vec![TupElement::Int(42), TupElement::String("anything".to_string())];
        assert_eq!(compare(&schema, &candidate), None);
    }

    #[test]
    fn test_compare_with_variable_type_constraint() {
        let schema = vec![
            TupElement::Int(42),
            TupElement::Variable(Variable::with_types(vec![ValueType::String])),
        ];
        let candidate1 = vec![TupElement::Int(42), TupElement::String("ok".to_string())];
        assert_eq!(compare(&schema, &candidate1), None);

        let candidate2 = vec![TupElement::Int(42), TupElement::Int(99)];
        assert_eq!(compare(&schema, &candidate2), Some(vec![1]));
    }

    #[test]
    fn test_compare_with_maybe_more() {
        let schema = vec![TupElement::Int(42), TupElement::MaybeMore];
        let candidate = vec![
            TupElement::Int(42),
            TupElement::String("extra".to_string()),
            TupElement::Bool(true),
        ];
        assert_eq!(compare(&schema, &candidate), None);
    }

    #[test]
    fn test_compare_without_maybe_more_length_mismatch() {
        let schema = vec![TupElement::Int(42)];
        let candidate = vec![TupElement::Int(42), TupElement::String("extra".to_string())];
        assert_eq!(compare(&schema, &candidate), Some(vec![1]));
    }

    #[test]
    fn test_compare_nested_tuples() {
        let schema = vec![
            TupElement::Int(1),
            TupElement::Tuple(vec![
                TupElement::String("nested".to_string()),
                TupElement::Variable(Variable::any()),
            ]),
        ];
        let candidate = vec![
            TupElement::Int(1),
            TupElement::Tuple(vec![
                TupElement::String("nested".to_string()),
                TupElement::Bool(true),
            ]),
        ];
        assert_eq!(compare(&schema, &candidate), None);
    }

    #[test]
    fn test_compare_nested_tuples_mismatch() {
        let schema = vec![
            TupElement::Int(1),
            TupElement::Tuple(vec![TupElement::String("nested".to_string()), TupElement::Int(99)]),
        ];
        let candidate = vec![
            TupElement::Int(1),
            TupElement::Tuple(vec![
                TupElement::String("nested".to_string()),
                TupElement::Int(100),
            ]),
        ];
        assert_eq!(compare(&schema, &candidate), Some(vec![1, 1]));
    }
}
