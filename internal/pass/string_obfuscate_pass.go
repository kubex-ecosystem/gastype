package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/rafa-mori/gastype/internal/astutil"
)

type StringObfuscatePass struct{}

func NewStringObfuscatePass() *StringObfuscatePass { return &StringObfuscatePass{} }
func (p *StringObfuscatePass) Name() string        { return "StringObfuscate" }

func (p *StringObfuscatePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	// Imports → ignorar
	importSpecs := make(map[*ast.BasicLit]bool)
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, spec := range gd.Specs {
				if is, ok := spec.(*ast.ImportSpec); ok && is.Path != nil {
					importSpecs[is.Path] = true
				}
			}
		}
	}

	// Constantes string
	constStrings := make(map[*ast.BasicLit]string)

	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}

		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 {
				continue
			}

			// Se não for string-like → ignora
			if !astutil.DetectStringLikeConst(vs, ctx) {
				continue
			}

			// Nome da constante
			name := vs.Names[0].Name

			// Se já tiver valor literal, guarda pra obfuscar
			if len(vs.Values) > 0 {
				if bl, ok := vs.Values[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
					constStrings[bl] = name
				}
			} else {
				// Caso não tenha valor literal (ex: alias sem inicialização), ainda é string-like
				// mas não vamos mexer agora
				fmt.Printf("  ℹ️ Found string alias constant without literal: %s\n", name)
			}
		}
	}

	// Tags de struct → ignorar
	structTags := make(map[*ast.BasicLit]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		if f, ok := n.(*ast.Field); ok && f.Tag != nil {
			structTags[f.Tag] = true
		}
		return true
	})

	var initStmts []ast.Stmt

	// Percorre a AST e obfusca strings
	ast.Inspect(file, func(n ast.Node) bool {

		bl, ok := n.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return true
		}
		if importSpecs[bl] || structTags[bl] {
			return true
		}

		val, err := strconv.Unquote(bl.Value)
		if err != nil || len(val) < 4 {
			return true
		}

		commonWords := []string{"main", "func", "package", "import", "var", "const", "if", "else", "for", "range"}
		for _, w := range commonWords {
			if val == w {
				return true
			}
		}

		// Obfuscação → byte array
		byteVals := make([]string, len(val))
		for i, b := range []byte(val) {
			byteVals[i] = strconv.Itoa(int(b))
		}

		// Caso seja constante → converte para var + init()
		if constName, isConst := constStrings[bl]; isConst {
			// Valida se é alias de string
			if !astutil.DetectStringLikeConst(&ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent(constName)},
				Values: []ast.Expr{},
				Type:   ast.NewIdent("string"),
			}, ctx) {
				if isConst {
					// Se for alias de string, não faz nada
					// bl.Value = constName
					// bl.Kind = token.IDENT
					return true
				} else {
					// Não é alias, converte para var
					// Substitui declaração const por var
					for _, decl := range file.Decls {
						if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.CONST {
							for i, spec := range gd.Specs {
								vs, ok := spec.(*ast.ValueSpec)
								if !ok || len(vs.Names) == 0 || vs.Names[0].Name != constName {
									continue
								}
								gd.Tok = token.VAR
								gd.Specs[i] = vs
								break
							}
						}
					}
				}
				// fmt.Printf("  ℹ️ Converted const to var: %s\n", constName)
			} else {
				// Gera const com array de bytes no init()
				initStmts = append(initStmts, &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(constName)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("string"),
							Args: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{Elt: ast.NewIdent("byte")},
									Elts: func() []ast.Expr {
										elts := []ast.Expr{}
										for _, b := range []byte(val) {
											elts = append(elts, &ast.BasicLit{
												Kind:  token.INT,
												Value: strconv.Itoa(int(b)),
											})
										}
										return elts
									}(),
								},
							},
						},
					},
				})
				transformations++
				return true
			}

		}

		if !bl.Kind.IsLiteral() {
			// Caso literal → inline
			bl.Value = fmt.Sprintf("string([]byte{%s})", strings.Join(byteVals, ", "))
			transformations++
			return true
		}

		return false
	})

	if len(initStmts) > 0 {
		file.Decls = append(file.Decls, &ast.FuncDecl{
			Name: ast.NewIdent("init"),
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{List: initStmts},
		})
	}

	if transformations > 0 {
		fmt.Printf("  🔄 StringObfuscatePass: %d transformations applied\n", transformations)
	}

	return nil
}
