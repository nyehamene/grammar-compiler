;; Tag queries for the grammar language

;; Definitions
(rule
  name: (ident) @definition.rule
)

(binding
  name: (ident) @definition.binding
)


(member_access
  object: (ident) @reference
)

(member_access
  property: (ident) @reference
)
