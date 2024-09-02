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
    scope: 'string',
    begin: /[^\/]/,
    end: /(?=[\/\(])/,
  };

  const DATA = {
    scope: 'number',
    begin: /[^\/,\(\)=<>\s]/,
    end: /(?=[\/,\(\)=<>%\s])/,
  };

  const VARIABLE = {
    scope: 'variable',
    begin: /</,
    end: />/,
    keywords: {
      $$pattern: /[^:|<>]+/,
      keyword: ['int', 'uint', 'bool', 'num', 'bint', 'str', 'bytes', 'uuid', 'tup'],
    },
  };

  const REFERENCE = {
    scope: 'reference',
    begin: /:/,
    end: /,/,
  };

  const TUPLE = {
    scope: 'tuple',
    begin: /\(/,
    end: /\)/,
    endsParent: true,
    keywords: {
      $$pattern: /[^,\)\s]+/,
      literal: ['nil', 'true', 'false'],
    },
    contains: [STRING, VARIABLE, REFERENCE, COMMENT, DATA, 'self'],
  };

  const DIRECTORY = {
    scope: 'directory',
    begin: /\//,
    end: /(?=\=)/,
    contains: [STRING, TUPLE, DSTRING],
  };

  const VALUE = {
    scope: 'value',
    begin: /=/,
    end: /\s/,
    keywords: {
      $$pattern: /[^=\s]+/,
      literal: ['nil', 'true', 'false'],
    },
    contains: [STRING, VARIABLE, REFERENCE, DATA],
  };

  hljs.registerLanguage('fql', (hljs) => ({
    classNameAliases: {
      directory: 'built_in',
      tuple: 'built_in',
      value: 'built_in',
      reference: 'variable',
      escape: 'subst',
    },
    contains: [DIRECTORY, TUPLE, VALUE, VARIABLE, COMMENT],
  }));
})();
