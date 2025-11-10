//! FQL query tokenizer/scanner
//!
//! This module tokenizes FQL query strings into tokens for the parser.

/// Token kinds recognized by the scanner
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TokenKind {
    Whitespace,
    Newline,
    Escape,
    Other,
    End,
    KeyValSep,   // =
    DirSep,      // /
    TupStart,    // (
    TupEnd,      // )
    TupSep,      // ,
    VarStart,    // <
    VarEnd,      // >
    VarSep,      // |
    StrMark,     // "
    StampStart,  // #
    StampSep,    // :
    Reserved,
}

/// A token with its kind and text
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Token {
    pub kind: TokenKind,
    pub text: String,
}

/// Scanner for FQL queries
pub struct Scanner {
    input: String,
    position: usize,
}

impl Scanner {
    pub fn new(input: impl Into<String>) -> Self {
        Self {
            input: input.into(),
            position: 0,
        }
    }

    /// Scan the next token
    pub fn scan(&mut self) -> Token {
        if self.position >= self.input.len() {
            return Token {
                kind: TokenKind::End,
                text: String::new(),
            };
        }

        let remaining = &self.input[self.position..];
        let ch = remaining.chars().next().unwrap();

        // Match single-character tokens
        let token = match ch {
            '=' => self.single_char_token(TokenKind::KeyValSep),
            '/' => self.single_char_token(TokenKind::DirSep),
            '(' => self.single_char_token(TokenKind::TupStart),
            ')' => self.single_char_token(TokenKind::TupEnd),
            ',' => self.single_char_token(TokenKind::TupSep),
            '<' => self.single_char_token(TokenKind::VarStart),
            '>' => self.single_char_token(TokenKind::VarEnd),
            '|' => self.single_char_token(TokenKind::VarSep),
            '"' => self.single_char_token(TokenKind::StrMark),
            '#' => self.single_char_token(TokenKind::StampStart),
            ':' => self.single_char_token(TokenKind::StampSep),
            '\\' => self.escape_token(),
            '\t' | ' ' => self.whitespace_token(),
            '\n' | '\r' => self.newline_token(),
            _ if is_reserved(ch) => self.single_char_token(TokenKind::Reserved),
            _ => self.other_token(),
        };

        token
    }

    fn single_char_token(&mut self, kind: TokenKind) -> Token {
        let ch = self.input[self.position..].chars().next().unwrap();
        self.position += ch.len_utf8();
        Token {
            kind,
            text: ch.to_string(),
        }
    }

    fn escape_token(&mut self) -> Token {
        let mut text = String::new();
        let chars: Vec<char> = self.input[self.position..].chars().collect();

        if chars.len() >= 2 {
            text.push(chars[0]); // \
            text.push(chars[1]); // escaped char
            self.position += chars[0].len_utf8() + chars[1].len_utf8();
        }

        Token {
            kind: TokenKind::Escape,
            text,
        }
    }

    fn whitespace_token(&mut self) -> Token {
        let start = self.position;
        while self.position < self.input.len() {
            match self.input[self.position..].chars().next().unwrap() {
                '\t' | ' ' => self.position += 1,
                _ => break,
            }
        }
        Token {
            kind: TokenKind::Whitespace,
            text: self.input[start..self.position].to_string(),
        }
    }

    fn newline_token(&mut self) -> Token {
        let start = self.position;
        while self.position < self.input.len() {
            match self.input[self.position..].chars().next().unwrap() {
                '\t' | ' ' | '\n' | '\r' => self.position += 1,
                _ => break,
            }
        }
        Token {
            kind: TokenKind::Newline,
            text: self.input[start..self.position].to_string(),
        }
    }

    fn other_token(&mut self) -> Token {
        let start = self.position;
        while self.position < self.input.len() {
            let ch = self.input[self.position..].chars().next().unwrap();
            if is_special(ch) || ch.is_whitespace() {
                break;
            }
            self.position += ch.len_utf8();
        }
        Token {
            kind: TokenKind::Other,
            text: self.input[start..self.position].to_string(),
        }
    }
}

fn is_special(ch: char) -> bool {
    matches!(
        ch,
        '=' | '/' | '(' | ')' | ',' | '<' | '>' | '|' | '"' | '#' | ':' | '\\'
    ) || is_reserved(ch)
}

fn is_reserved(ch: char) -> bool {
    matches!(
        ch,
        '!' | '$' | '%' | '&' | '{' | '}' | '*' | '+' | ';' | '?' | '@' | '[' | ']' | '^' | '`'
            | '~'
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_scanner_basic() {
        let mut scanner = Scanner::new("/dir/(1,2)=42");
        assert_eq!(scanner.scan().kind, TokenKind::DirSep);
        assert_eq!(scanner.scan().kind, TokenKind::Other); // "dir"
        assert_eq!(scanner.scan().kind, TokenKind::DirSep);
        assert_eq!(scanner.scan().kind, TokenKind::TupStart);
    }
}
