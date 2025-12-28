;; Tag queries for the grammar language

;; Definitions
(rule_declaration
  name: (ident) @definition.rule
)

(binding
  name: (ident) @definition.binding
)

;; References
(non_terminal
  (ident) @reference
)

(member_access
  object: (ident) @reference
)

(member_access
  property: (ident) @reference
)
