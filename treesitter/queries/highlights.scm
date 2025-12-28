;; Highlights for the grammar language

;; Keywords
"@import" @keyword

;; Operators and punctuation
"=" @operator
"|" @operator
"." @operator ; for member access
";" @punctuation.delimiter
"(" @punctuation.delimiter
")" @punctuation.delimiter
"[" @punctuation.bracket
"]" @punctuation.bracket
"{" @punctuation.bracket
"}" @punctuation.bracket

;; Rule names (as fields, as requested)
(rule_declaration
  name: (ident) @field)

;; Binding names (as variables, as requested)
(binding
  name: (ident) @variable)

;; Non-terminal references
(non_terminal (ident) @variable)

;; Member access identifiers
(member_access
  object: (ident) @variable)
(member_access
  property: (ident) @variable)

;; Literals
(string_literal) @string
(regex_literal) @string.regexp

;; Comments
(comment) @comment
