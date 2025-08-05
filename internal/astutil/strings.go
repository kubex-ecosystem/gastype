package astutil

import (
	"go/ast"
	"go/token"
	"go/types"
	"unicode"
)

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
