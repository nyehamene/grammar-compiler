;; Indentation queries for the grammar language, based on the Helix editor
;; documentation and the project's formatting rules.

;; Indent content inside multiline groups.
;; The #not-one-line? predicate ensures this only applies
;; when the group spans multiple lines.
(group
  _ @indent
  (#not-one-line? @parent))

(optional
  _ @indent
  (#not-one-line? @parent))

(repetition
  _ @indent
  (#not-one-line? @parent))

;; Align '|' in vertical alternatives with the '=' of the rule.
;; This creates the aligned vertical layout for alternatives as described
;; in the formatting rules.
(rule
  "=" @anchor
  (production
    "|" @align
    (#not-same-line? @anchor @align)))

;; Align trailing ';' in vertical rules with the '=' of the rule.
;; This ensures the semicolon at the end of a multiline rule is
;; also aligned with the '='.
(rule
  "=" @anchor
  ";" @align
  (#not-same-line? @anchor @align))
