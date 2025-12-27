package ast

import "grammar/token"

// Visitor defines the interface for an AST visitor.
// The parent node is passed to the visit function.
type Visitor func(node, parent Node)

// Walk traverses an AST in depth-first order.
func Walk(node Node, visitor Visitor) {
	walk(node, nil, visitor)
}

func walk(node, parent Node, visitor Visitor) {
	if node == nil {
		return
	}

	visitor(node, parent)

	switch n := node.(type) {
	case *File:
		for _, decl := range n.Decls {
			walk(decl, n, visitor)
		}
	case *RuleDecl:
		walk(n.Name, n, visitor)
		if n.Body != nil {
			for _, expr := range n.Body {
				walk(expr, n, visitor)
			}
		}
	case *BindingDecl:
		walk(n.Name, n, visitor)
		walk(n.Path, n, visitor)
	case *AlternativeExpr:
		for _, expr := range n.Exprs {
			walk(expr, n, visitor)
		}
	case *SequenceExpr:
		for _, expr := range n.Exprs {
			walk(expr, n, visitor)
		}
	case *OptionalExpr:
		walk(n.Expr, n, visitor)
	case *RepetitionExpr:
		walk(n.Expr, n, visitor)
	case *GroupExpr:
		walk(n.Expr, n, visitor)
	case *MemberExpr:
		walk(n.Object, n, visitor)
		walk(n.Member, n, visitor)
	case *DirectiveExpr:
		walk(n.Name, n, visitor)
		for _, arg := range n.Args {
			walk(arg, n, visitor)
		}
	case *CommentGroup:
		for _, c := range n.List {
			walk(c, n, visitor)
		}
		// Terminals (Ident, StringLit, RegexLit, Comment) have no children to walk.
	}
}

// FindNodeAt finds the most specific AST node at a given position.
func FindNodeAt(file *File, pos token.Pos) (node Node, parent Node) {
	Walk(file, func(n, p Node) {
		if n == nil {
			return
		}
		// If the node's range contains the position
		if pos >= n.Pos() && pos < n.End() {
			// If we haven't found a node yet, or if this node is more specific
			// (i.e., its range is smaller than the last one found), update it.
			if node == nil || (n.Pos() >= node.Pos() && n.End() <= node.End()) {
				node = n
				parent = p
			}
		}
	})
	return
}
