//! Streaming support for range queries
//!
//! Provides async streaming of query results.

use futures::stream::Stream;
use keyval::KeyValue;

/// Create a stream of key-values from a range query
pub fn stream_range(
    _results: Vec<KeyValue>,
) -> impl Stream<Item = Result<KeyValue, crate::EngineError>> {
    // TODO: Implement actual async streaming with FoundationDB
    futures::stream::iter(vec![])
}
