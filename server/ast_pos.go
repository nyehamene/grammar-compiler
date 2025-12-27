package server

import (
	"grammar/ast"
	"grammar/token"
)

// FindNodeAt finds the most specific AST node at a given position, and its parent.
func FindNodeAt(file *ast.File, pos token.Pos) (ast.Node, ast.Node) {
	var found, parent ast.Node
	var visit func(node, p ast.Node)

	visit = func(node, p ast.Node) {
		if node == nil {
			return
		}

		if pos >= node.Pos() && pos < node.End() {
			if found == nil || (node.Pos() >= found.Pos() && node.End() <= found.End()) {
				found = node
				parent = p
			}

			switch n := node.(type) {
			case *ast.File:
				for _, decl := range n.Decls {
					visit(decl, n)
				}
			case *ast.RuleDecl:
				visit(n.Name, n)
				if n.Body != nil {
					for _, expr := range n.Body {
						visit(expr, n)
					}
				}
			case *ast.BindingDecl:
				visit(n.Name, n)
				visit(n.Path, n)
			case *ast.AlternativeExpr:
				for _, expr := range n.Exprs {
					visit(expr, n)
				}
			case *ast.SequenceExpr:
				for _, expr := range n.Exprs {
					visit(expr, n)
				}
			case *ast.OptionalExpr:
				visit(n.Expr, n)
			case *ast.RepetitionExpr:
				visit(n.Expr, n)
			case *ast.GroupExpr:
				visit(n.Expr, n)
			case *ast.MemberExpr:
				visit(n.Object, n)
				visit(n.Member, n)
			case *ast.DirectiveExpr:
				visit(n.Name, n)
				for _, arg := range n.Args {
					visit(arg, n)
				}
			case *ast.CommentGroup:
				for _, c := range n.List {
					visit(c, n)
				}
			}
		}
	}

	visit(file, nil)
	return found, parent
}
