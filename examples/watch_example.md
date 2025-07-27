# Watch Feature Example

This document demonstrates how to use the new `--watch` flag to monitor key-value changes in FoundationDB using FQL.

## Basic Usage

To watch a single key for changes:

```bash
fql --cluster my_cluster_file --watch --query '/my/dir("key1")=<>'
```

This will:
1. Print the current value of the key
2. Monitor the key for any changes
3. Print the new value whenever it changes
4. Continue monitoring until interrupted (Ctrl+C)

## Examples

### Watch a specific key:
```bash
fql --watch -q '/users("john")=<name:string>'
```

### Watch with value type constraints:
```bash
fql --watch -q '/counters("page_views")=<count:int>'
```

### Watch with multiple tuple elements:
```bash
fql --watch -q '/events("user123", "login")=<timestamp:int>'
```

## Limitations

- Watch only works with single-read queries (queries that target a specific key)
- Cannot be used with range queries (`/dir(..)=<>`)
- Cannot be used with write operations (`/dir("key")=value`)
- Cannot be used with directory queries (`/dir`)
- Only one query can be watched at a time

## Error Examples

These will produce errors:

```bash
# Error: Multiple queries not supported
fql --watch -q '/key1=<>' -q '/key2=<>'

# Error: Range queries not supported  
fql --watch -q '/dir(..)=<>'

# Error: Write operations not supported
fql --watch -q '/dir("key")=42'

# Error: Directory queries not supported
fql --watch -q '/dir'
```

## Use Cases

- Monitoring configuration changes
- Tracking counter updates
- Watching for user actions
- Debugging data flow issues
- Real-time data monitoring