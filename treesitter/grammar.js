module.exports = grammar({
  name: 'grammar',

  rules: {
    source: $ => repeat($.declaration),

    declaration: $ => choice(
      $.binding,
      $.rule,
      $.comment
    ),

    binding: $ => seq(
      field('name', $.ident),
      '=',
      field('value', $.import_directive),
      ';'
    ),

    rule: $ => seq(
      field('name', $.ident),
      '=',
      field('body', $.production),
      ';'
    ),

    comment: $ => token(seq('//', /[^\n]*/, optional("\n"))),

    production: $ => prec.left(1, choice(
      seq($.production, '|', $.sequence),
      $.sequence
    )),

    sequence: $ => repeat1($._basic),

    _basic: $ => choice(
      $.group,
      $.repetition,
      $.optional,
      $._term
    ),

    group: $ => seq('(', field('content', $.production), ')'),
    repetition: $ => seq('{', field('content', $.production), '}'),
    optional: $ => seq('[', field('content', $.production), ']'),

    _term: $ => choice(
        $._terminal,
        $._non_terminal
    ),

    _non_terminal: $ => choice(
      $.ident,
      $.member_access
    ),

    _terminal: $ => choice(
      $.string,
      $.regexp
    ),

    member_access: $ => prec.left(2, seq(
      field('object', choice($.ident, $.member_access)),
      '.',
      field('property', $.ident)
    )),

    import_directive: $ => seq(
      '@import',
      '(' ,
      field('path', $.string),
      ')'
    ),

    string: $ => seq(
      '"',
      repeat(choice(
        $.string_content,
        $.escape_sequence
      )),
      '"'
    ),
    string_content: $ => token.immediate(/[^"\\n\\]+/),

    regexp: $ => seq(
      '/',
      repeat(choice(
        $.regexp_content,
        $.escape_sequence
      )),
      '/'
    ),
    regexp_content: $ => token.immediate(/[^/\\]+/),

    escape_sequence: $ => token.immediate(/\\./),

    ident: $ => token(/[a-zA-Z_][a-zA-Z0-9_]*/)
  }
});
