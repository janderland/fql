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
      /(kb|mb|gb)?/, // allow a byte-unit to appear after numbers.
    ],
    beginScope: {
      1: 'accent',
      2: 'number',
      3: 'accent',
      4: 'number',
      5: 'accent',
      6: 'number',
      7: 'accent',
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

  const VSTAMP = {
    begin: [
      /#/,
      /[A-Fa-f0-9]*/,
      /:/,
      /[A-Fa-f0-9]{4}/,
    ],
    beginScope: {
      1: 'accent',
      2: 'number',
      3: 'accent',
      4: 'number',
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
      'remove',
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
      'vstamp',
      'append',
      'sum',
      'avg',
      'min',
      'max',
      'count',
      'be',
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
      'bigendian',
      'width',
      'unsigned',
      'reverse',
      'limit',
      'mode',
      'want_all',
      'iterator',
      'exact',
      'small',
      'medium',
      'large',
      'serial',
      'snapshot',
      'strict',
      'rand',
      'pick',
    ].join(' '),
  };

  const OPTIONS = {
    scope: 'options',
    begin: /\[/,
    end: /]/,
    keywords: {
      $$pattern: /[^,:"]+/,
      keyword: [
        'be',
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
        'endian',
        'width',
        'unsigned',
        'reverse',
        'limit',
        'mode',
        'snapshot',
        'strict',
        'pick',
        'sep',
      ],
    },
    contains: [
      STRING,
      {
        begin: [
          /:/,
          /[^,\]"]+/,
        ],
        beginScope: {
          1: 'option',
          2: 'number',
        },
      },
    ],
  };

  const INLINEOPT = {
    scope: 'options',
    begin: /(?=\b(?:width|unsigned)\b)/,
    end: /(?=\s|$)/,
    keywords: OPTIONS.keywords,
    contains: OPTIONS.contains,
  };

  const VAR_NAME = {
    begin: [
      /[\w\.]+/,
      /:/,
    ],
    beginScope: {
      1: 'number',
      2: 'variable',
    },
  };

  const VARIABLE = {
    begin: /</,
    beginScope: 'variable',
    end: />/,
    endScope: 'variable',
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
        'vstamp',
        'append',
        'sum',
        'avg',
        'min',
        'max',
        'count',
      ],
    },
    contains: [
      VAR_NAME,
      OPTIONS,
      {
        scope: 'variable',
        begin: /\|/,
      },
    ],
  };

  const REF_TYPE = {
    begin: [
      /!/,
      /\w+/,
    ],
    beginScope: {
      1: 'reference',
      2: 'keyword',
    },
  };

  const REFERENCE = {
    begin: [
      /:/,
      /[\w\.]+/,
    ],
    beginScope: {
      1: 'reference',
      2: 'number',
    },
    contains: [REF_TYPE],
  };

  const MAYBEMORE = {
    scope: 'variable',
    begin: /\.\.\./,
  };

  const DIRECTORY = {
    scope: 'directory',
    begin: /\//,
    end: /(?=[\(=\s]|$)/,
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
      VSTAMP,
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
      VSTAMP,
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
      REFERENCE,
      MAYBEMORE,
      INLINEOPT,
      KEYWORD,
      STRING,
      UUID,
      VSTAMP,
      BYTES,
      NUMBER,
      OPTIONS,
      { // Highlight lone bar for inline text.
        scope: 'variable',
        begin: /\|/,
      },
    ],
  }));
})();
