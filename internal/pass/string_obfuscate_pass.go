package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/rafa-mori/gastype/internal/astutil"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

type objectSpecMap struct {
	isConst         bool
	name            string
	obfuscatedName  string
	varType         ast.Expr
	obfuscatedValue any
}

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
	objectsSpecMap := make(map[any]objectSpecMap)

	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 {
				continue
			}

			// Nome do objeto
			name := vs.Names[0].Name
			// Tipo do objeto
			varType := vs.Type

			// Se já tiver valor literal, guarda pra obfuscar
			if len(vs.Values) > 0 {
				ast.Inspect(vs.Values[0], func(n ast.Node) bool {
					bl, ok := n.(*ast.BasicLit)
					if !ok || bl.Kind != token.STRING {
						return true
					}
					objectsSpecMap[bl] = objectSpecMap{
						isConst:         gd.Tok == token.CONST,
						name:            name,
						varType:         varType,
						obfuscatedName:  fmt.Sprintf("obfuscated_%s", name),
						obfuscatedValue: nil,
					}
					return false
				})
			} else {
				objectsSpecMap[vs.Names[0]] = objectSpecMap{
					isConst:         gd.Tok == token.CONST,
					name:            name,
					varType:         varType,
					obfuscatedName:  fmt.Sprintf("obfuscated_%s", name),
					obfuscatedValue: nil,
				}
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

	// var initStmts []ast.Stmt

	// Percorre a AST e obfusca strings
	ast.Inspect(file, func(n ast.Node) bool {

		bl, ok := n.(*ast.BasicLit)

		if !ok || bl.Kind != token.STRING {
			return true
		}

		if importSpecs[bl] || structTags[bl] {
			return true
		}

		if len(bl.Value) < 2 {
			return true
		}

		// Ignora strings curtas demais e comuns
		// Ex: "", "a", "ok", "id", "to", "is", "in", "on", "at", "of", "by"
		if len(bl.Value) <= 3 {
			simpleStrings := map[string]bool{
				`""`: true, `"a"`: true, `"b"`: true, `"c"`: true,
				`"d"`: true, `"e"`: true, `"f"`: true, `"g"`: true,
				`"h"`: true, `"i"`: true, `"j"`: true, `"k"`: true,
				`"l"`: true, `"m"`: true, `"n"`: true, `"o"`: true,
				`"p"`: true, `"q"`: true, `"r"`: true, `"s"`: true,
				`"t"`: true, `"u"`: true, `"v"`: true, `"w"`: true,
				`"x"`: true, `"y"`: true, `"z"`: true,
				`"ok"`: true, `"id"`: true, `"to"`: true,
				`"is"`: true, `"in"`: true, `"on"`: true,
				`"at"`: true, `"of"`: true, `"by"`: true,
			}
			if simpleStrings[bl.Value] {
				return true
			}
		}

		if astutil.CheckConstant(
			bl,
			bl.Kind.String(),
			ctx,
		) {
			// Se for constante, não obfusca
			return true
		}

		// Valor da strings

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
		isConst := objectsSpecMap[bl].isConst
		isConstCheck := astutil.CheckConstant(
			bl,
			bl.Kind.String(),
			ctx,
		)

		if !isConst && !isConstCheck {
			if !astutil.CheckConstant(
				bl,
				bl.Kind.String(),
				ctx,
			) {
				tpS := objectsSpecMap[bl].varType
				tp := "string"
				if tpS != nil {
					switch t := tpS.(type) {
					case *ast.Ident:
						tp = t.Name
					case *ast.SelectorExpr:
						if x, ok := t.X.(*ast.Ident); ok {
							tp = fmt.Sprintf("%s.%s", x.Name, t.Sel.Name)
						}
					}
				}

				// Se for outro tipo (alias), mantém o tipo
				// Substitui o literal pela conversão do byte array
				// Ex: "hello" → string([]byte{104, 101, 108, 108, 111})
				// Ex: MyStringAlias("hello") → MyStringAlias([]byte{104, 101, 108, 108, 111})
				bl.Value = fmt.Sprintf("%s([]byte{%s})", tp, strings.Join(byteVals, ", "))
				transformations++
			}

			transformations++

			return true
		}

		return false
	})

	if transformations > 0 {
		gl.Log("info", fmt.Sprintf("🔄 StringObfuscatePass: %d transformations applied", transformations))
	}

	return nil
}
