package astutil

import (
	"go/ast"
	"go/token"
)

// MenorTipoParaFlags retorna o tipo de inteiro necessário para armazenar as flags
func MenorTipoParaFlags(numFlags int) string {
	switch {
	case numFlags <= 8:
		return "uint8"
	case numFlags <= 16:
		return "uint16"
	case numFlags <= 32:
		return "uint32"
	default:
		return "uint64"
	}
}

// GetConstNames returns the constant names for fields
func GetConstNames(fields []string) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = "Flag" + field
	}
	return names
}

// InferExprType infers the type of an expression
func InferExprType(expr ast.Expr, ctx *TranspileContext) (ast.Expr, string, bool) {
	typeInfer := func(expr ast.Expr) (ast.Expr, string, bool) {
		switch v := expr.(type) {
		case *ast.BasicLit:
			switch v.Kind {
			case token.INT:
				return &ast.Ident{Name: "int"}, "int", true
			case token.FLOAT:
				return &ast.Ident{Name: "float64"}, "float64", true
			case token.CHAR:
				return &ast.Ident{Name: "rune"}, "rune", true
			case token.STRING:
				return &ast.Ident{Name: "string"}, "string", true
			}
		case *ast.Ident:
			if v.Obj != nil && v.Obj.Kind == ast.Con {
				if spec, ok := v.Obj.Decl.(*ast.ValueSpec); ok {
					if len(spec.Values) > 0 {
						return InferExprType(spec.Values[0], ctx)
					}
				}
			}
			if v.Obj != nil && v.Obj.Kind == ast.Var {
				if spec, ok := v.Obj.Decl.(*ast.ValueSpec); ok {
					if spec.Type != nil {
						return spec.Type, spec.Type.(*ast.Ident).Name, true
					}
					if len(spec.Values) > 0 {
						return InferExprType(spec.Values[0], ctx)
					}
				}
			}
		case *ast.CompositeLit:
			return v.Type, v.Type.(*ast.Ident).Name, true
		case *ast.CallExpr:
			if funIdent, ok := v.Fun.(*ast.Ident); ok {
				if funIdent.Obj != nil && funIdent.Obj.Kind == ast.Fun {
					if funcDecl, ok := funIdent.Obj.Decl.(*ast.FuncDecl); ok {
						if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
							return funcDecl.Type.Results.List[0].Type, funcDecl.Type.Results.List[0].Type.(*ast.Ident).Name, true
						}
					}
				}
			}
		case *ast.BinaryExpr:
			leftType, _, _ := InferExprType(v.X, ctx)
			rightType, _, _ := InferExprType(v.Y, ctx)
			if leftType != nil && rightType != nil {
				if leftIdent, ok := leftType.(*ast.Ident); ok {
					if rightIdent, ok := rightType.(*ast.Ident); ok {
						if leftIdent.Name == rightIdent.Name {
							return leftType, leftIdent.Name, true
						}
					}
				}
			}
		case *ast.ParenExpr:
			return InferExprType(v.X, ctx)
		case *ast.SelectorExpr:
			return v.Sel, v.Sel.Name, true
		}
		return nil, "", false
	}

	if expr != nil {
		if inferredType, inferredName, ok := typeInfer(expr); ok {
			return inferredType, inferredName, ok
		}
		return nil, "", false
	}

	return nil, "", false
}

func GetTypeName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.StarExpr:
		return GetTypeName(v.X)
	case *ast.ArrayType:
		return "[]" + GetTypeName(v.Elt)
	case *ast.MapType:
		return "map[" + GetTypeName(v.Key) + "]" + GetTypeName(v.Value)
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan " + GetTypeName(v.Value)
	}
	return ""
}

// CheckVariable checks if a variable with the given name and type exists in the context
func CheckVariable(varName string, varType string, ctx *TranspileContext) bool {
	currentNode := ast.NewIdent(varName)

	ast.Inspect(currentNode, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok {
			for i, name := range vs.Names {
				if name.Name == varName {
					if vs.Type != nil {
						if ident, ok := vs.Type.(*ast.Ident); ok {
							if ident.Name == varType {

								return false
							}
						}
					} else if len(vs.Values) > i {
						inferredType, _, _ := InferExprType(vs.Values[i], ctx)
						if inferredType != nil {
							if ident, ok := inferredType.(*ast.Ident); ok {
								if ident.Name == varType {

									return false
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	return false
}

// GetParentNode returns the parent node of a given node in the AST
func GetParentNode(file *ast.File, node ast.Node) ast.Node {
	var parent ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == node {
			return false
		}
		if n.Pos() < node.Pos() && n.End() > node.Pos() {
			parent = n
		}
		return true
	})
	return parent
}
