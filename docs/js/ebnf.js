(function() {
  // EBNF (Extended Backus-Naur Form) syntax highlighting
  // Based on ISO/IEC 14977 standard

  const COMMENT = {
    scope: 'comment',
    begin: /\(\*/,
    end: /\*\)/,
  };

  const TERMINAL_SINGLE = {
    scope: 'string',
    begin: /'/,
    end: /'/,
  };

  const TERMINAL_DOUBLE = {
    scope: 'string',
    begin: /"/,
    end: /"/,
  };

  const SPECIAL_SEQUENCE = {
    scope: 'meta',
    begin: /\?/,
    end: /\?/,
  };

  const RULE_NAME = {
    // Rule definition: name followed by =
    begin: [
      /[a-zA-Z_][\w\-]*/,
      /\s*/,
      /=/,
    ],
    beginScope: {
      1: 'title.function',
      3: 'operator',
    },
  };

  const RULE_REFERENCE = {
    // Reference to another rule (non-terminal)
    scope: 'variable',
    begin: /[a-zA-Z_][\w\-]*/,
  };

  const REPETITION = {
    // Curly braces for repetition
    scope: 'punctuation',
    begin: /[{}]/,
  };

  const OPTIONAL = {
    // Square brackets for optional
    scope: 'punctuation',
    begin: /[\[\]]/,
  };

  const GROUPING = {
    // Parentheses for grouping
    scope: 'punctuation',
    begin: /[()]/,
  };

  const OPERATORS = {
    scope: 'operator',
    begin: /[|,;\-]/,
  };

  const NUMBER = {
    scope: 'number',
    begin: /\b\d+\b/,
  };

  hljs.registerLanguage('ebnf', (_hljs) => ({
    name: 'EBNF',
    case_insensitive: false,
    contains: [
      COMMENT,
      TERMINAL_SINGLE,
      TERMINAL_DOUBLE,
      SPECIAL_SEQUENCE,
      RULE_NAME,
      NUMBER,
      OPERATORS,
      REPETITION,
      OPTIONAL,
      GROUPING,
      RULE_REFERENCE,
    ],
  }));
})();
