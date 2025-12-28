module.exports = grammar({
  name: 'grammar',

  rules: {
    source: $ => repeat($.declaration),

    declaration: $ => choice(
      $.binding,
      $.rule_declaration,
      $.comment
    ),

    binding: $ => seq(
      field('name', $.ident),
      '=',
      field('value', $.import_directive),
      ';'
    ),

    rule_declaration: $ => seq(
      field('name', $.ident),
      '=',
      field('body', $.rule_body_expression),
      ';'
    ),

    comment: $ => token(seq('//', /[^\n]*/)),

    rule_body_expression: $ => prec.left(
      1, // Lower precedence for '|' than concatenation
      choice(
        seq($.rule_body_expression, '|', $.sequence_expression),
        $.sequence_expression
      )
    ),

    sequence_expression: $ => repeat1($._basic_production_unit),

    _basic_production_unit: $ => choice(
      $.optional_production,
      $.repetition_production,
      $.group_production,
      $.term_production
    ),

    optional_production: $ => seq('[', field('content', $.rule_body_expression), ']'),
    repetition_production: $ => seq('{', field('content', $.rule_body_expression), '}'),
    group_production: $ => seq('(', field('content', $.rule_body_expression), ')'),

    term_production: $ => repeat1($.basic),

    basic: $ => choice(
      $.terminal,
      $.non_terminal
    ),

    terminal: $ => choice(
      $.string_literal,
      $.regex_literal
    ),

    non_terminal: $ => choice(
      $.ident,
      $.member_access
    ),

    member_access: $ => prec.left(2, seq( // Higher precedence than '|' and concatenation
      field('object', choice($.ident, $.member_access)),
      '.',
      field('property', $.ident)
    )),

    import_directive: $ => seq(
      '@import',
      '(' ,
      field('path', $.string_literal),
      ')'
    ),

    string_literal: $ => token(seq('"', /[^"\n]*/, '"')),
    ident: $ => token(/[a-zA-Z_][a-zA-Z0-9_]*/),
    regex_literal: $ => token(seq('/', /[^\/]*/, '/'))
  }
});
