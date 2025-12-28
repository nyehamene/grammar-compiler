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
(rule
  name: (ident) @variable.other.member)

;; Binding names (as variables, as requested)
(binding
  name: (ident) @namespace)


;; Member access identifiers
(member_access
  object: (ident) @variable)
(member_access
  property: (ident) @variable)

;; Literals
(string) @string
(regexp) @string.regexp

;; Comments
(comment) @comment
