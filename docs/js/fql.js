(function() {
  const ESCAPE = {
    scope: 'escape',
    begin: /\\/,
    end: /./,
  };

  const COMMENT = {
    scope: 'comment',
    begin: /%/,
    end: /\n/,
  };

  const STRING = {
    scope: 'string',
    begin: /"/,
    end: /"/,
    contains: [ESCAPE],
  };

  const DSTRING = {
    scope: 'section',
    begin: /[^\/]/,
    end: /(?=[\/\(])/,
  };

  const NUMBER = {
    scope: 'number',
    begin: /[^\/,\(\)=<>\s]/,
    end: /(?=[\/,\(\)=<>%\s])/,
    contains: [{
      scope: 'title',
      begin: /\./,
    }],
  };

  const COMPOUND = {
    scope: 'title',
    begin: /(#|-)/,
    contains: [NUMBER],
  };

  const BYTES = {
    begin: [
      /0/,
      /x/,
    ],
    beginScope: {
      1: 'number',
      2: 'title',
    },
    contains: [NUMBER],
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
      $$pattern: /[^:|<>]+/,
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
    begin: /:/,
    end: /,/,
  };

  const MORE = {
    scope: 'variable',
    begin: /\.\.\./,
  };

  const TUPLE = {
    scope: 'tuple',
    begin: /\(/,
    end: /\)/,
    endsParent: true,
    contains: [COMMENT, STRING, VARIABLE, REFERENCE, MORE, KEYWORD, UUID, BYTES, COMPOUND, NUMBER, 'self'],
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
    contains: [TUPLE, STRING, VARIABLE, REFERENCE, KEYWORD, UUID, BYTES, COMPOUND, NUMBER],
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
    contains: [DIRECTORY, G_TUPLE, VALUE, VARIABLE, MORE, KEYWORD, COMMENT, STRING, UUID, BYTES, COMPOUND, NUMBER],
  }));
})();
