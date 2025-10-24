package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/kubex-ecosystem/gastype/internal/astutil"
	stdastutil "golang.org/x/tools/go/ast/astutil"

	gl "github.com/kubex-ecosystem/logz/logger"
)

// BoolToFlagsPass converte campos bool em flags bitwise
type BoolToFlagsPass struct{}

func NewBoolToFlagsPass() *BoolToFlagsPass {
	return &BoolToFlagsPass{}
}

func (p *BoolToFlagsPass) Name() string {
	return "BoolToFlags"
}

func (p *BoolToFlagsPass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	// Guarda mapeamento struct → campos booleanos convertidos
	convertedStructs := make(map[string][]string)

	// === 1️⃣ Identifica e converte structs com bools ===
	ast.Inspect(file, func(n ast.Node) bool {
		typeDecl, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeDecl.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structName := typeDecl.Name.Name
		boolFields := []string{}

		// Usa IInfo para checar tipo real
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue
			}
			if tv, ok := ctx.GetTypes()[field.Type]; ok && tv.Type.String() == "bool" {
				for _, name := range field.Names {
					boolFields = append(boolFields, name.Name)
				}
			}
		}

		if !astutil.DeveConverterBools(len(boolFields)) {
			return true
		}

		convertedStructs[structName] = boolFields

		// Determina tipo ideal (uint8, uint16, uint32, uint64)
		flagType := astutil.MenorTipoParaFlags(len(boolFields))

		// === 2️⃣ Cria constantes AST reais ===
		constDecls := []ast.Decl{}
		for i, fieldName := range boolFields {
			constName := fmt.Sprintf("Flag%s_%s", structName, fieldName)

			constSpec := &ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(constName)},
				Type:  ast.NewIdent(flagType),
				Values: []ast.Expr{
					&ast.BinaryExpr{
						X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
						Op: token.SHL,
						Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", i)},
					},
				},
			}

			constDecls = append(constDecls, &ast.GenDecl{
				Tok:   token.CONST,
				Specs: []ast.Spec{constSpec},
			})

			ctx.AddDef(ast.NewIdent(constName), nil)
			gl.Log("info", fmt.Sprintf("Added constant: %s (%s)", constName, flagType))
		}
		file.Decls = append(constDecls, file.Decls...) // insere no topo

		// === 3️⃣ Substitui campo bool por "flags" ===
		newFields := []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("flags")},
				Type:  ast.NewIdent(flagType),
			},
		}
		for _, field := range structType.Fields.List {
			if tv, ok := ctx.GetTypes()[field.Type]; ok && tv.Type.String() == "bool" {
				continue
			}
			newFields = append(newFields, field)
		}
		structType.Fields.List = newFields

		return true
	})

	// === 4️⃣ Substitui acessos cfg.Debug → cfg.flags & FlagStruct_Debug != 0 ===
	stdastutil.Apply(file, func(cr *stdastutil.Cursor) bool {
		sel, ok := cr.Node().(*ast.SelectorExpr)
		if !ok {
			return true
		}

		selInfo := ctx.GetSelections()[sel]
		if selInfo == nil || selInfo.Obj() == nil {
			return true
		}

		fieldName := selInfo.Obj().Name()
		if fieldName == "" {
			return true
		}

		// Verifica se o campo pertence a uma struct que foi convertida
		for structName, boolFields := range convertedStructs {
			if contains(boolFields, fieldName) {
				recvType := selInfo.Recv().String()
				if recvType == structName || recvType == "*"+structName {
					// Substituir o acesso pelo bitwise check
					cr.Replace(&ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  &ast.SelectorExpr{X: sel.X, Sel: ast.NewIdent("flags")},
							Op: token.AND,
							Y:  ast.NewIdent(fmt.Sprintf("Flag%s_%s", structName, fieldName)),
						},
						Op: token.NEQ,
						Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
					})
				}
			}
		}

		return true
	}, nil)

	return nil
}

// contains verifica se s está em slice arr
func contains(arr []string, s string) bool {
	for _, v := range arr {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
