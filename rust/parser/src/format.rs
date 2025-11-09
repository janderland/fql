//! Query formatting module
//!
//! Converts keyval structures back into FQL query strings.

use keyval::*;

/// Format a query as an FQL query string
pub fn format(query: &Query) -> String {
    match query {
        Query::KeyValue(kv) => format_keyvalue(kv),
        Query::Key(key) => format_key(key),
        Query::Directory(dir) => format_directory(dir),
    }
}

fn format_keyvalue(kv: &KeyValue) -> String {
    format!("{}={}", format_key(&kv.key), format_value(&kv.value))
}

fn format_key(key: &Key) -> String {
    format!("{}{}", format_directory(&key.directory), format_tuple(&key.tuple))
}

fn format_directory(dir: &Directory) -> String {
    if dir.is_empty() {
        return String::new();
    }
    dir.iter()
        .map(|elem| match elem {
            DirElement::String(s) => format!("/{}", s),
            DirElement::Variable(var) => format!("/<{}>", format_variable(var)),
        })
        .collect()
}

fn format_tuple(tup: &Tuple) -> String {
    let elements: Vec<String> = tup.iter().map(format_tup_element).collect();
    format!("({})", elements.join(","))
}

fn format_tup_element(elem: &TupElement) -> String {
    match elem {
        TupElement::Tuple(t) => format_tuple(t),
        TupElement::Nil => "nil".to_string(),
        TupElement::Int(i) => i.to_string(),
        TupElement::Uint(u) => u.to_string(),
        TupElement::Bool(b) => b.to_string(),
        TupElement::Float(f) => f.to_string(),
        TupElement::String(s) => format!("\"{}\"", s),
        TupElement::Uuid(u) => u.to_string(),
        TupElement::Bytes(b) => format!("0x{}", hex::encode(b)),
        TupElement::Variable(var) => format!("<{}>", format_variable(var)),
        TupElement::MaybeMore => "...".to_string(),
        TupElement::VStamp(_) => "#vstamp".to_string(),
        TupElement::VStampFuture(_) => "#vstamp_future".to_string(),
    }
}

fn format_value(val: &Value) -> String {
    match val {
        Value::Tuple(t) => format_tuple(t),
        Value::Nil => "nil".to_string(),
        Value::Int(i) => i.to_string(),
        Value::Uint(u) => u.to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Float(f) => f.to_string(),
        Value::String(s) => format!("\"{}\"", s),
        Value::Uuid(u) => u.to_string(),
        Value::Bytes(b) => format!("0x{}", hex::encode(b)),
        Value::Variable(var) => format!("<{}>", format_variable(var)),
        Value::Clear => "clear".to_string(),
        Value::VStamp(_) => "#vstamp".to_string(),
        Value::VStampFuture(_) => "#vstamp_future".to_string(),
    }
}

fn format_variable(var: &Variable) -> String {
    if var.types.is_empty() {
        String::new()
    } else {
        var.types.iter()
            .map(|t| t.to_string())
            .collect::<Vec<_>>()
            .join("|")
    }
}

// Add hex dependency placeholder
mod hex {
    pub fn encode(_data: &[u8]) -> String {
        // TODO: Use actual hex encoding crate
        String::from("...")
    }
}
