;; Text objects for the grammar language

;;; General purpose text objects

(comment) @comment.outer

;; Note: .inner for strings and regex includes the delimiters because
;; the grammar structure is not known. A more precise grammar might
;; have a specific "content" node to hook into.
(string) @string.outer
(string) @string.inner

(regexp) @regex.outer
(regexp) @regex.inner


;;; Language-specific text objects

;; A whole declaration (rule or binding)
(declaration) @declaration.outer

;; The body of a rule
(rule
  body: (_) @rule.inner)
(rule) @rule.outer

;; The value of a binding
(binding
  value: (_) @binding.inner)
(binding) @binding.outer

;; Content of a group (...)
(group (production) @group.inner)
(group) @group.outer

;; Content of an optional [...]
(optional (production) @optional.inner)
(optional) @optional.outer

;; Content of a repetition {...}
(repetition (production) @repetition.inner)
(repetition) @repetition.outer

;; The argument of an import
(import_directive (string) @import.inner)
(import_directive) @import.outer

;; A single production in a list of productions
(production (sequence) @item.inner)
(optional (production) @item.inner)
(repetition (production) @item.inner)

;; A whole production
(production) @production.outer
