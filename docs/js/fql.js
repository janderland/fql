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
    ],
    beginScope: {
      1: 'accent',
      2: 'number',
      3: 'accent',
      4: 'number',
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
    begin: /[\w\.]/,
  };

  const KEYWORD = {
    scope: 'keyword',
    beginKeywords: [
      'true',
      'false',
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
    ].join(' '),
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
    ],
  };

  hljs.registerLanguage('fql', (hljs) => ({
    classNameAliases: {
      directory: 'built_in',
      tuple: 'built_in',
      value: 'built_in',
      reference: 'variable',
      escape: 'subst',
      accent: 'title',
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
      { // Highlight lone bar for inline text.
        scope: 'variable',
        begin: /\|/,
      },
    ],
  }));
})();
