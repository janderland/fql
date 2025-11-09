//! Database facade/abstraction layer
//!
//! Provides traits for abstracting FoundationDB operations for testing.

use async_trait::async_trait;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum FacadeError {
    #[error("Transaction error: {0}")]
    TransactionError(String),
}

/// Trait for database transactions
#[async_trait]
pub trait Transaction {
    async fn get(&self, key: &[u8]) -> Result<Option<Vec<u8>>, FacadeError>;
    async fn set(&self, key: &[u8], value: &[u8]) -> Result<(), FacadeError>;
    async fn clear(&self, key: &[u8]) -> Result<(), FacadeError>;
    async fn get_range(&self, begin: &[u8], end: &[u8]) -> Result<Vec<(Vec<u8>, Vec<u8>)>, FacadeError>;
}

/// Trait for creating transactions
#[async_trait]
pub trait Database {
    type Transaction: Transaction;
    
    async fn create_transaction(&self) -> Result<Self::Transaction, FacadeError>;
}
