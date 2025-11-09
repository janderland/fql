//! FQL query parser
//!
//! This crate provides parsing functionality for FQL query strings,
//! converting them into the keyval data structures.

pub mod scanner;
pub mod format;

use keyval::*;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ParseError {
    #[error("Unexpected token: {0}")]
    UnexpectedToken(String),

    #[error("Invalid syntax: {0}")]
    InvalidSyntax(String),

    #[error("Scan error: {0}")]
    ScanError(String),
}

/// Parse an FQL query string into a Query structure.
pub fn parse(input: &str) -> Result<Query, ParseError> {
    // TODO: Implement full parser state machine
    // This is a simplified placeholder that demonstrates the API

    // Example: Parse a simple key-value query like "/dir/(1,2)=42"
    if input.is_empty() {
        return Err(ParseError::InvalidSyntax("Empty input".to_string()));
    }

    // For now, return a placeholder
    Ok(Query::KeyValue(KeyValue {
        key: Key {
            directory: vec![],
            tuple: vec![TupElement::Int(1)],
        },
        value: Value::Int(42),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_placeholder() {
        let result = parse("/test");
        assert!(result.is_ok());
    }
}
