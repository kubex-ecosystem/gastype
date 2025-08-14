// Package pass implements various AST transformation passes for the Gastype project.
package pass

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// FieldAccessToBitwisePass: troca usos “cfg.Debug” em args/returns/exprs por bitwise check
type FieldAccessToBitwisePass struct{}

func NewFieldAccessToBitwisePass() *FieldAccessToBitwisePass { return &FieldAccessToBitwisePass{} }
func (p *FieldAccessToBitwisePass) Name() string             { return "FieldAccessToBitwise" }

func (p *FieldAccessToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	// helper interno para transformar um expr se for selector de campo bool convertido
	transform := func(expr ast.Expr) ast.Expr {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok {
			return nil
		}
		selInfo := ctx.GetSelections()[sel]
		if selInfo == nil || selInfo.Obj() == nil || selInfo.Recv() == nil {
			return nil
		}
		fieldName := selInfo.Obj().Name()
		typName := selInfo.Recv().String()
		if strings.HasPrefix(typName, "*") {
			typName = typName[1:]
		}
		if info, exists := ctx.Structs[typName]; exists {
			if flagName, ok := info.FlagMapping[fieldName]; ok {
				transformations++
				return &ast.BinaryExpr{
					X: &ast.ParenExpr{
						X: &ast.BinaryExpr{
							X: &ast.SelectorExpr{X: sel.X, Sel: ast.NewIdent("flags")},
							Op: token.AND,
							Y:  ast.NewIdent(flagName),
						},
					},
					Op: token.NEQ,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
				}
			}
		}
		return nil
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			for i, a := range node.Args {
				if r := transform(a); r != nil {
					node.Args[i] = r
				}
			}
		case *ast.ReturnStmt:
			for i, r := range node.Results {
				if rr := transform(r); rr != nil {
					node.Results[i] = rr
				}
			}
		case *ast.AssignStmt:
			for i, r := range node.Rhs {
				if rr := transform(r); rr != nil {
					node.Rhs[i] = rr
				}
			}
		case *ast.ExprStmt:
			if rr := transform(node.X); rr != nil {
				node.X = rr
			}
		}
		return true
	})

	if transformations > 0 {
		ctx.LogVerbose(fset, "🔄 FieldAccessToBitwisePass: %d transformations applied", transformations)
	}
	return nil
}

