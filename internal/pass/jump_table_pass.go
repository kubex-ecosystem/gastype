// Package pass implements AST transformation passes.
package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	stdastutil "golang.org/x/tools/go/ast/astutil"

	"github.com/kubex-ecosystem/gastype/internal/astutil"
)

// JumpTablePass: if/else encadeado por igualdade da MESMA variável string → jump table.
type JumpTablePass struct{}

func NewJumpTablePass() *JumpTablePass { return &JumpTablePass{} }
func (p *JumpTablePass) Name() string  { return "JumpTable" }

func (p *JumpTablePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	stdastutil.Apply(file, func(c *stdastutil.Cursor) bool {
		ifStmt, ok := c.Node().(*ast.IfStmt)
		if !ok {
			return true
		}

		varName, branches := p.collectBranches(ifStmt)
		if varName == "" || len(branches) < 3 {
			return true
		}
		transformations++

		jumpTableName := fmt.Sprintf("jumpTable_%s", varName)

		// var jumpTable_varName = map[string]func(){ "a": func(){...}, ... }
		mapEntries := []ast.Expr{}
		for _, b := range branches {
			mapEntries = append(mapEntries, &ast.KeyValueExpr{
				Key:   &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(b.matchValue)},
				Value: &ast.FuncLit{Type: &ast.FuncType{Params: &ast.FieldList{}}, Body: b.body},
			})
		}
		mapDecl := &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(jumpTableName)},
						Type: &ast.MapType{
							Key:   ast.NewIdent("string"),
							Value: &ast.FuncType{Params: &ast.FieldList{}},
						},
						Values: []ast.Expr{
							&ast.CompositeLit{
								Type: &ast.MapType{
									Key:   ast.NewIdent("string"),
									Value: &ast.FuncType{Params: &ast.FieldList{}},
								},
								Elts: mapEntries,
							},
						},
					},
				},
			},
		}

		// if fn, ok := jumpTable[varName]; ok { fn() }
		lookup := &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("fn"), ast.NewIdent("ok")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.IndexExpr{X: ast.NewIdent(jumpTableName), Index: ast.NewIdent(varName)},
				},
			},
			Cond: ast.NewIdent("ok"),
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("fn")}}}},
		}

		c.Replace(&ast.BlockStmt{List: []ast.Stmt{mapDecl, lookup}})
		return true
	}, nil)

	if transformations > 0 {
		ctx.LogVerbose(nil, "🔀 JumpTablePass: %d transformations applied", transformations)
	}
	return nil
}

type branchInfo struct {
	matchValue string
	body       *ast.BlockStmt
}

func (p *JumpTablePass) collectBranches(ifStmt *ast.IfStmt) (string, []branchInfo) {
	branches := []branchInfo{}
	var varName string

	cur := ifStmt
	for cur != nil {
		bin, ok := cur.Cond.(*ast.BinaryExpr)
		if !ok || bin.Op != token.EQL {
			break
		}
		left, ok := bin.X.(*ast.Ident)
		right, ok2 := bin.Y.(*ast.BasicLit)
		if !ok || !ok2 || right.Kind != token.STRING {
			break
		}
		if varName == "" {
			varName = left.Name
		} else if left.Name != varName {
			break
		}
		branches = append(branches, branchInfo{
			matchValue: right.Value[1 : len(right.Value)-1], // tira aspas
			body:       cur.Body,
		})
		if next, ok := cur.Else.(*ast.IfStmt); ok {
			cur = next
		} else {
			break
		}
	}
	return varName, branches
}
