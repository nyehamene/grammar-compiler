package ast

import "grammar/token"

// Node represents a node in the abstract syntax tree.
type Node interface {
	Pos() token.Pos // Position of first character belonging to the node
	End() token.Pos // Position of first character immediately after the node
}

// Expr represents an expression node in the AST.
type Expr interface {
	Node
	exprNode()
}

// Decl represents a declaration node in the AST.
type Decl interface {
	Node
	declNode()
}

// File represents a grammar file.
type File struct {
	Decls  []Decl
	EndPos token.Pos
}

// RuleDecl represents a rule declaration.
type RuleDecl struct {
	Name   *Ident
	Body   []Expr // { production }
	EndPos token.Pos
}

// BindingDecl represents a binding declaration.
type BindingDecl struct {
	Name   *Ident
	Path   *StringLit
	EndPos token.Pos
}

// Ident represents an identifier.
type Ident struct {
	NamePos token.Pos
	Name    string
}

// StringLit represents a string literal.
type StringLit struct {
	ValuePos token.Pos
	Value    string
}

// RegexLit represents a regular expression literal.

type RegexLit struct {
	ValuePos token.Pos

	Value string
}

// ExternalValue represents a reference to an external value, like $foo.

type ExternalValue struct {
	DollarPos token.Pos // position of "$"

	Name string // identifier name

}

// AlternativeExpr represents a sequence of alternative productions.

type AlternativeExpr struct {
	Exprs []Expr // production { "|" production }

}

// OptionalExpr represents an optional production.

type OptionalExpr struct {
	Lbrack token.Pos

	Expr Expr

	Rbrack token.Pos
}

// RepetitionExpr represents a repeated production.

type RepetitionExpr struct {
	Lbrace token.Pos

	Expr Expr

	Rbrace token.Pos
}

// GroupExpr represents a grouped production.

type GroupExpr struct {
	Lparen token.Pos

	Expr Expr

	Rparen token.Pos
}

// TermExpr represents a basic term in a production.

type TermExpr struct {
	X Expr // terminal, non_terminal, or directive

}

// DirectiveExpr represents a directive.

type DirectiveExpr struct {
	AtPos token.Pos

	Name *Ident

	Args []Expr

	EndPos token.Pos
}

// Comment represents a single comment.

type Comment struct {
	Slash token.Pos

	Text string
}

// CommentGroup represents a sequence of comments.

type CommentGroup struct {
	List []*Comment
}

// SequenceExpr represents a sequence of expressions.

type SequenceExpr struct {
	Exprs []Expr
}

// MemberExpr represents a member access expression.

type MemberExpr struct {
	Object Expr

	Member *Ident
}

func (f *File) Pos() token.Pos { return token.NoPos } // A file does not have a specific position

func (f *File) End() token.Pos { return f.EndPos }

func (r *RuleDecl) Pos() token.Pos { return r.Name.Pos() }

func (r *RuleDecl) End() token.Pos { return r.EndPos }

func (b *BindingDecl) Pos() token.Pos { return b.Name.Pos() }

func (b *BindingDecl) End() token.Pos { return b.EndPos }

func (i *Ident) Pos() token.Pos { return i.NamePos }

func (i *Ident) End() token.Pos { return token.Pos(int(i.NamePos) + len(i.Name)) }

func (s *StringLit) Pos() token.Pos { return s.ValuePos }

func (s *StringLit) End() token.Pos { return token.Pos(int(s.ValuePos) + len(s.Value)) }

func (r *RegexLit) Pos() token.Pos { return r.ValuePos }

func (r *RegexLit) End() token.Pos { return token.Pos(int(r.ValuePos) + len(r.Value)) }

func (e *ExternalValue) Pos() token.Pos { return e.DollarPos }

func (e *ExternalValue) End() token.Pos { return token.Pos(int(e.DollarPos) + 1 + len(e.Name)) }

func (a *AlternativeExpr) Pos() token.Pos { return a.Exprs[0].Pos() }

func (a *AlternativeExpr) End() token.Pos { return a.Exprs[len(a.Exprs)-1].End() }

func (s *SequenceExpr) Pos() token.Pos { return s.Exprs[0].Pos() }

func (s *SequenceExpr) End() token.Pos { return s.Exprs[len(s.Exprs)-1].End() }

func (o *OptionalExpr) Pos() token.Pos { return o.Lbrack }

func (o *OptionalExpr) End() token.Pos { return o.Rbrack + 1 }

func (r *RepetitionExpr) Pos() token.Pos { return r.Lbrace }

func (r *RepetitionExpr) End() token.Pos { return r.Rbrace + 1 }

func (g *GroupExpr) Pos() token.Pos { return g.Lparen }

func (g *GroupExpr) End() token.Pos { return g.Rparen + 1 }

func (t *TermExpr) Pos() token.Pos { return t.X.Pos() }

func (t *TermExpr) End() token.Pos { return t.X.End() }

func (d *DirectiveExpr) Pos() token.Pos { return d.AtPos }

func (d *DirectiveExpr) End() token.Pos { return d.EndPos }

func (m *MemberExpr) Pos() token.Pos { return m.Object.Pos() }

func (m *MemberExpr) End() token.Pos { return m.Member.End() }

func (*RuleDecl) declNode() {}

func (*BindingDecl) declNode() {}

func (*Ident) exprNode() {}

func (*StringLit) exprNode() {}

func (*RegexLit) exprNode() {}

func (*ExternalValue) exprNode() {}

func (*AlternativeExpr) exprNode() {}

func (*SequenceExpr) exprNode() {}

func (*OptionalExpr) exprNode() {}

func (*RepetitionExpr) exprNode() {}

func (*GroupExpr) exprNode() {}

func (*TermExpr) exprNode() {}

func (*DirectiveExpr) exprNode() {}

func (*MemberExpr) exprNode() {}

func (c *Comment) Pos() token.Pos { return c.Slash }

func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

func (cg *CommentGroup) Pos() token.Pos { return cg.List[0].Pos() }

func (cg *CommentGroup) End() token.Pos { return cg.List[len(cg.List)-1].End() }

func (*CommentGroup) declNode() {}
