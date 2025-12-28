;; Defines local variables and their scopes for the grammar language.

;; Rule and binding definitions
(rule
  name: (ident) @definition.var)

(binding
  name: (ident) @definition.var)


(member_access
  object: (ident) @reference.var)

(member_access
  property: (ident) @reference.var)
