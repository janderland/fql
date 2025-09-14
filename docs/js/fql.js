(function() {
  const ESCAPE = {
    scope: 'escape',
    begin: /\\./,
  };

  const COMMENT = {
    scope: 'comment',
    begin: /%.*\n?/,
  };

  const NUMBER = {
    begin: [
      /-?/,
      /\d+/,
      /\.?/,
      /\d*/,
      /e?/,
      /\d*/,
    ],
    beginScope: {
      1: 'accent',
      2: 'number',
      3: 'accent',
      4: 'number',
      5: 'accent',
      6: 'number',
    },
  };

  const BYTES = {
    begin: [
      /0/,
      /x/,
      /[A-Za-z0-9]*/,
    ],
    beginScope: {
      1: 'number',
      2: 'accent',
      3: 'number',
    },
  };

  const UUID = {
    begin: [
      /\w{8}/,
      /-/,
      /\w{4}/,
      /-/,
      /\w{4}/,
      /-/,
      /\w{4}/,
      /-/,
      /\w{12}/,
    ],
    beginScope: {
      1: 'number',
      2: 'accent',
      3: 'number',
      4: 'accent',
      5: 'number',
      6: 'accent',
      7: 'number',
      8: 'accent',
      9: 'number',
    },
  };

  const STRING = {
    scope: 'string',
    begin: /"/,
    end: /"/,
    contains: [ESCAPE],
  };

  const DSTRING = {
    scope: 'section',
    begin: /[\w\.\-]/,
  };

  const KEYWORD = {
    scope: 'keyword',
    beginKeywords: [
      'true',
      'false',
      '-inf',
      'inf',
      '-nan',
      'nan',
      'clear',
      'nil',
      'any',
      'int',
      'bool',
      'num',
      'bint',
      'str',
      'bytes',
      'uuid',
      'tup',
      'append',
      'sum',
      'count',
      'be',
      'le',
      'i8',
      'i16',
      'i32',
      'i64',
      'u8',
      'u16',
      'u32',
      'u64',
      'f32',
      'f64',
      'f80',
      'reverse',
      'limit',
    ].join(' '),
  };

  const OPTIONS = {
    scope: 'options',
    begin: /\[/,
    end: /]/,
    keywords: {
      $$pattern: /[^,:]+/,
      keyword: [
        'be',
        'le',
        'i8',
        'i16',
        'i32',
        'i64',
        'u8',
        'u16',
        'u32',
        'u64',
        'f32',
        'f64',
        'f80',
        'reverse',
        'limit',
      ],
    },
    contains: [
      {
        begin: [
          /:/,
          /[^,\]]/
        ],
        beginScope: {
          1: 'option',
          2: 'number',
        },
      },
    ],
  };

  const VARIABLE = {
    scope: 'variable',
    begin: /</,
    end: />/,
    keywords: {
      $$pattern: /[^:|]+/,
      keyword: [
        'any',
        'int',
        'bool',
        'num',
        'bint',
        'str',
        'bytes',
        'uuid',
        'tup',
        'append',
        'sum',
        'count',
      ],
    },
    contains: [
      OPTIONS,
    ],
  };

  const REFERENCE = {
    scope: 'reference',
    begin: /:[\w\.]+/,
  };

  const MAYBEMORE = {
    scope: 'variable',
    begin: /\.\.\./,
  };

  const DIRECTORY = {
    scope: 'directory',
    begin: /\//,
    end: /(?=\()/,
    contains: [
      STRING,
      VARIABLE,
      DSTRING,
    ],
  };

  const TUPLE = {
    scope: 'tuple',
    begin: /\(/,
    end: /\)/,
    contains: [
      COMMENT,
      'self',
      STRING,
      VARIABLE,
      REFERENCE,
      MAYBEMORE,
      KEYWORD,
      UUID,
      BYTES,
      NUMBER,
      OPTIONS,
    ],
  };

  const VALUE = {
    scope: 'value',
    begin: /=/,
    end: /[\s%]/,
    contains: [
      TUPLE,
      STRING,
      VARIABLE,
      REFERENCE,
      KEYWORD,
      UUID,
      BYTES,
      NUMBER,
      OPTIONS,
    ],
  };

  hljs.registerLanguage('fql', (_hljs) => ({
    classNameAliases: {
      directory: 'built_in',
      tuple: 'built_in',
      value: 'built_in',
      reference: 'variable',
      escape: 'subst',
      accent: 'title',
      options: 'title',
    },
    contains: [
      COMMENT, 
      DIRECTORY,
      TUPLE,
      VALUE,
      VARIABLE,
      MAYBEMORE,
      KEYWORD,
      STRING,
      UUID,
      BYTES,
      NUMBER,
      OPTIONS,
      { // Highlight lone bar & semicolon for inline text.
        scope: 'variable',
        begin: /\||:/,
      },
    ],
  }));
})();
