/*! highlightjs-fql | Apache-2.0 | https://github.com/janderland/fql */
var hljsFQL = (function (hljs) {
  'use strict';

  /*
  Language: FQL
  Description: FoundationDB Query Language
  Website: https://github.com/janderland/fql
  Category: database
  */

  // Keyword categories
  const LITERALS    = ['true', 'false', 'nil', 'inf', '-inf', 'nan', '-nan'];
  const VERBS       = ['clear', 'remove'];
  const TYPES       = ['any', 'int', 'bool', 'num', 'str', 'bytes', 'uuid', 'tup', 'vstamp'];
  const TYPES_INT   = ['i8', 'i16', 'i32', 'i64', 'u8', 'u16', 'u32', 'u64'];
  const TYPES_FLOAT = ['f32', 'f64', 'f80'];
  const AGGREGATES  = ['append', 'sum', 'avg', 'min', 'max', 'count'];
  const OPTIONS_KW  = [
    'be', 'bigendian', 'unsigned', 'width', 'raw',
    'reverse', 'snapshot', 'limit', 'mode',
    'separator',
    'strict',
  ];
  const OPTION_VALUES  = ['wantall', 'iterator', 'exact', 'small', 'medium', 'large', 'serial'];

  // US-keyboard ASCII punctuation. Used to construct broad terminator regexes
  // (e.g. OPTION.end). Excludes `_` since it's a word character.
  // Pre-escaped for use inside a regex `[...]` char class.
  const SYMBOLS = '`~!@#$%^&*()\\-=+[\\]{}\\\\|;:\'",.<>/?';

  // Build a lookahead-end regex: triggers on whitespace, EOL, or any keyboard
  // symbol not matched by `keep`. Used by OPTION/DIRECTORY/VALUE/TYPE/TYPE_CAST.
  const endAtSymbol = (keep) =>
    new RegExp('(?=[\\s' + SYMBOLS.replace(keep, '') + ']|$)');

  // Build a multi-capture numeric mode where scopes strictly alternate between
  // 'accent' and 'number'. `firstIsAccent` picks the starting phase.
  //
  // Example: NUMBER is built with firstIsAccent=true and the parts below.
  // For input `-42.3e7`:
  //   parts[0] /-?/   capture 1  accent   matches '-'
  //   parts[1] /\d+/  capture 2  number   matches '42'
  //   parts[2] /\.?/  capture 3  accent   matches '.'
  //   parts[3] /\d*/  capture 4  number   matches '3'
  //   parts[4] /e?/   capture 5  accent   matches 'e'
  //   parts[5] /\d*/  capture 6  number   matches '7'
  const numericMode = (parts, firstIsAccent) => ({
    begin: parts,
    beginScope: Object.fromEntries(
      parts.map((_, i) => [i + 1, (i % 2 === 0) === firstIsAccent ? 'accent' : 'number']),
    ),
  });

  // Build a beginKeywords mode scoped 'keyword'. Used by KEYWORD/OPTNAME.
  const keywordMode = (words) => ({
    scope: 'keyword',
    beginKeywords: words.join(' '),
  });

  // Composed lists for specific modes.
  const TOP_KEYWORDS = [
    ...LITERALS, ...VERBS, ...TYPES, ...TYPES_INT, ...TYPES_FLOAT,
    ...AGGREGATES, ...OPTIONS_KW, ...OPTION_VALUES,
  ];
  const VARIABLE_KEYWORDS = [...LITERALS, ...TYPES, ...AGGREGATES];
  const BRACKET_KEYWORDS  = [...OPTIONS_KW, ...TYPES_INT, ...TYPES_FLOAT];

  function fql(_hljs) {
    const ESCAPE = {
      scope: 'escape',
      begin: /\\./,
    };

    const COMMENT = {
      scope: 'comment',
      begin: /%.*\n?/,
    };

    const NUMBER = numericMode(
      [/-?/, /\d+/, /\.?/, /\d*/, /e?/, /\d*/],
      true,
    );

    const BYTES = numericMode(
      [/0/, /x/, /[A-Za-z0-9]*/],
      false,
    );

    const UUID = numericMode(
      [/\w{8}/, /-/, /\w{4}/, /-/, /\w{4}/, /-/, /\w{4}/, /-/, /\w{12}/],
      false,
    );

    const VSTAMP = numericMode(
      [/#/, /[A-Fa-f0-9]*/, /:/, /[A-Fa-f0-9]{4}/],
      true,
    );

    const STRING = {
      scope: 'string',
      begin: /"/,
      end: /"/,
      contains: [ESCAPE],
    };

    const DSTRING = {
      scope: 'dstring',
      begin: /[\w\.\-]/,
    };

    const KEYWORD = keywordMode(TOP_KEYWORDS);

    // OPTNAME — recognizes a token as an option-keyword name.
    const OPTNAME = keywordMode(OPTIONS_KW);

    // OPTION — a single option, with or without `:value`. The single entry
    // point for option syntax everywhere — inside [...], at top level, in
    // TUPLE, in VALUE. Ends at any keyboard symbol other than `:`, plus
    // whitespace and EOL.
    const OPTION = {
      scope: 'accent',
      begin: new RegExp('(?=\\b(?:' + OPTIONS_KW.join('|') + ')\\b)'),
      end: endAtSymbol(/:/g),
      contains: [
        OPTNAME,
        STRING,
        {
          begin: [
            /:/,
            /[\w.\-]+/,
          ],
          beginScope: {
            1: 'option',
            2: 'number',
          },
        },
      ],
    };

    const OPTIONS = {
      scope: 'options',
      begin: /\[/,
      end: /]/,
      keywords: {
        $$pattern: /[^,:"]+/,
        keyword: BRACKET_KEYWORDS,
      },
      contains: [
        OPTION,
        STRING,
      ],
    };

    const VAR_NAME = {
      begin: [
        /[\w\.]+/,
        /:/,
      ],
      beginScope: {
        1: 'params',
        2: 'variable',
      },
    };

    // TYPE — a type-name keyword optionally followed by [options]. Used inside
    // VARIABLE (per `|`-separated type) so options bind to the immediately
    // preceding type rather than the variable as a whole. beginKeywords
    // auto-scopes the keyword itself, so no outer `scope:` is needed (and
    // adding one would double-wrap the keyword span).
    const TYPE = {
      beginKeywords: VARIABLE_KEYWORDS.join(' '),
      end: endAtSymbol(/\[/g),
      contains: [OPTIONS],
    };

    const VARIABLE = {
      begin: /</,
      beginScope: 'variable',
      end: />/,
      endScope: 'variable',
      contains: [
        VAR_NAME,
        TYPE,
        {
          scope: 'variable',
          begin: /\|/,
        },
      ],
    };

    const TYPE_CAST = {
      begin: [
        /!/,
        /\w+/,
      ],
      beginScope: {
        1: 'reference',
        2: 'keyword',
      },
      end: endAtSymbol(/\[/g),
      contains: [OPTIONS],
    };

    const REFERENCE = {
      begin: [
        /:/,
        /[\w\.]+/,
      ],
      beginScope: {
        1: 'reference',
        2: 'params',
      },
    };

    const MAYBEMORE = {
      scope: 'variable',
      begin: /\.\.\./,
    };

    const DIRECTORY = {
      scope: 'directory',
      begin: /[/@]/,
      end: endAtSymbol(/[/@.\-"]/g),
      contains: [
        STRING,
        // Only the bare '<>' placeholder is a valid directory segment; any
        // other '<...>' terminates DIRECTORY and is parsed as a top-level
        // VARIABLE.
        { scope: 'variable', begin: /<>/ },
        MAYBEMORE,
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
        OPTION,
        REFERENCE,
        TYPE_CAST,
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
      end: endAtSymbol(/[(<\[:!#\-"]/g),
      contains: [
        TUPLE,
        STRING,
        VARIABLE,
        OPTION,
        REFERENCE,
        TYPE_CAST,
        KEYWORD,
        UUID,
        VSTAMP,
        BYTES,
        NUMBER,
        OPTIONS,
      ],
    };

    return {
      name: 'FQL',
      aliases: ['fql'],
      classNameAliases: {
        directory: 'built_in',
        tuple: 'built_in',
        value: 'built_in',

        reference: 'variable',
        dstring: 'section',

        options: 'title',
        accent: 'title',
        escape: 'subst',
      },
      contains: [
        COMMENT,
        DIRECTORY,
        TUPLE,
        VALUE,
        VARIABLE,
        OPTION,
        REFERENCE,
        TYPE_CAST,
        MAYBEMORE,
        KEYWORD,
        STRING,
        UUID,
        VSTAMP,
        BYTES,
        NUMBER,
        OPTIONS,
      ],
    };
  }

  hljs.registerLanguage('fql', fql);

  return fql;

})(hljs);
