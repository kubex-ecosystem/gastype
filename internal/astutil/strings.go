package astutil

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"unicode"
)

// CheckConstant verifica se um nó AST é uma constante de um tipo específico
func CheckConstant(node ast.Expr, expectedType string, ctx *TranspileContext) bool {
	tv, exists := ctx.GetTypes()[node] // Pega informações de tipo do ASTExpr [4]
	if !exists || !tv.Assignable() {   // Verifica se existe e se é uma constante [go/types]
		return false
	}

	// Verifica o tipo específico
	switch expectedType {
	case "string":
		return tv.Type.Underlying().String() == "string" || tv.Value.Kind() == constant.String
	case "int":
		return tv.Type.Underlying().String() == "int" || tv.Value.Kind() == constant.Int
	case "bool":
		return tv.Type.Underlying().String() == "bool" || tv.Value.Kind() == constant.Bool
	// Adicione outros tipos conforme necessário
	default:
		return false
	}
}

func DetectStringLikeConst(vs *ast.ValueSpec, ctx *TranspileContext) bool {
	if vs.Type != nil {
		if tv, ok := ctx.GetTypes()[vs.Type]; ok && tv.Type != nil {
			// Verifica se é alias de string
			if named, ok := tv.Type.(*types.Named); ok {
				if named.String() != "string" && tv.Type.Underlying().String() == "string" {
					return true
				}
			}
			// Ou string puro
			if tv.Type.Underlying().String() == "string" {
				return true
			}
		}
	}

	// Se não tem tipo explícito, mas tem valor literal string
	if len(vs.Values) > 0 {
		if bl, ok := vs.Values[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
			return false
		}
	}

	return false
}

func IsAllCaps(name string) bool {
	if name == "" {
		return false
	}
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func IsAllLower(name string) bool {
	if name == "" {
		return false
	}
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}

func IsConstant(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}

func IsIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}

func IsVar(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}
