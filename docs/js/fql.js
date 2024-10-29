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
      1: 'title',
      2: 'number',
      3: 'title',
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
      2: 'title',
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
      2: 'title',
      3: 'number',
      4: 'title',
      5: 'number',
      6: 'title',
      7: 'number',
      8: 'title',
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
      'int',
      'uint',
      'bool',
      'num',
      'bint',
      'str',
      'bytes',
      'uuid',
      'tup',
      'agg',
      'sum',
    ].join(' '),
  };

  const VARIABLE = {
    scope: 'variable',
    begin: /</,
    end: />/,
    keywords: {
      $$pattern: /[^:|]+/,
      keyword: [
        'int',
        'uint',
        'bool',
        'num',
        'bint',
        'str',
        'bytes',
        'uuid',
        'tup',
        'agg',
        'sum',
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

  const TUPLE = {
    scope: 'tuple',
    begin: /\(/,
    end: /\)/,
    endsParent: true,
    contains: [COMMENT, STRING, VARIABLE, REFERENCE, MAYBEMORE, KEYWORD, UUID, BYTES, NUMBER, 'self'],
  };

  const DIRECTORY = {
    scope: 'directory',
    begin: /\//,
    end: /(?=\=)/,
    contains: [STRING, VARIABLE, TUPLE, DSTRING],
  };

  const VALUE = {
    scope: 'value',
    begin: /=/,
    end: /[\s%]/,
    contains: [TUPLE, STRING, VARIABLE, REFERENCE, KEYWORD, UUID, BYTES, NUMBER],
  };

  // TODO: Refactor into single tuple.
  // We need this because TUPLE has
  // endsParent=true which doesn't
  // allow it to match a lone tuple.
  const G_TUPLE = Object.assign({}, TUPLE);
  G_TUPLE.endsParent = false;

  hljs.registerLanguage('fql', (hljs) => ({
    classNameAliases: {
      directory: 'built_in',
      tuple: 'built_in',
      value: 'built_in',
      reference: 'variable',
      escape: 'subst',
    },
    contains: [DIRECTORY, G_TUPLE, VALUE, VARIABLE, MAYBEMORE, KEYWORD, COMMENT, STRING, UUID, BYTES, NUMBER],
  }));
})();
