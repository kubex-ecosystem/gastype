// Package pass implements AST transformation passes.
package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// JumpTablePass detects AND transforms if-else chains comparing the same variable to strings
// into an optimized jump table lookup.
type JumpTablePass struct{}

func NewJumpTablePass() *JumpTablePass {
	return &JumpTablePass{}
}

func (p *JumpTablePass) Name() string {
	return "JumpTable"
}

func (p *JumpTablePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	// Precisamos manipular a AST com substituição de nós
	astutil.ReplaceNode(file, func(n ast.Node) ast.Node {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return n
		}

		varName, branches := p.collectBranches(ifStmt)
		if varName == "" || len(branches) < 3 {
			return n // não é candidato
		}

		transformations++

		// Nome único pro jump table
		jumpTableName := fmt.Sprintf("jumpTable_%s", varName)

		// Cria o map[string]func()
		mapEntries := []ast.Expr{}

		for _, b := range branches {
			mapEntries = append(mapEntries, &ast.KeyValueExpr{
				Key: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(b.matchValue),
				},
				Value: &ast.FuncLit{
					Type: &ast.FuncType{
						Params:  &ast.FieldList{},
						Results: nil,
					},
					Body: b.body,
				},
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
		lookupIf := &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent("fn"),
					ast.NewIdent("ok"),
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.IndexExpr{
						X:     ast.NewIdent(jumpTableName),
						Index: ast.NewIdent(varName),
					},
				},
			},
			Cond: ast.NewIdent("ok"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: ast.NewIdent("fn"),
						},
					},
				},
			},
		}

		// Substitui o if original por: var jumpTable... + if fn,ok...
		return &ast.BlockStmt{
			List: []ast.Stmt{
				mapDecl,
				lookupIf,
			},
		}

	}(file), nil)

	if transformations > 0 {
		ctx.LogVerbose(nil, "🔀 JumpTablePass: %d transformations applied", transformations)
	}

	return nil
}

// branchInfo guarda cada case do if-chain
type branchInfo struct {
	matchValue string
	body       *ast.BlockStmt
}

// collectBranches lê a cadeia if-else-if e retorna o nome da var e todos os cases
func (p *JumpTablePass) collectBranches(ifStmt *ast.IfStmt) (string, []branchInfo) {
	branches := []branchInfo{}
	var varName string

	current := ifStmt
	for current != nil {
		binExpr, ok := current.Cond.(*ast.BinaryExpr)
		if !ok || binExpr.Op != token.EQL {
			break
		}

		leftIdent, ok := binExpr.X.(*ast.Ident)
		if !ok {
			break
		}
		if varName == "" {
			varName = leftIdent.Name
		} else if leftIdent.Name != varName {
			break
		}

		rightLit, ok := binExpr.Y.(*ast.BasicLit)
		if !ok || rightLit.Kind != token.STRING {
			break
		}

		branches = append(branches, branchInfo{
			matchValue: rightLit.Value[1 : len(rightLit.Value)-1], // remove aspas
			body:       current.Body,
		})

		if nextIf, ok := current.Else.(*ast.IfStmt); ok {
			current = nextIf
		} else {
			break
		}
	}

	return varName, branches
}
