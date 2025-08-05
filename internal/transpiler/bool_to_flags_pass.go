// Package transpiler provides pass-based AST transformations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
)

// BoolToFlagsPass transforms boolean struct fields to bitwise flags
type BoolToFlagsPass struct{}

// Name returns the name of this pass
func (p *BoolToFlagsPass) Name() string {
	return "BoolToFlags"
}

// Apply executes the bool-to-flags transformation on the AST
func (p *BoolToFlagsPass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
	structsFound := 0
	fieldsTransformed := 0

	// First pass: Find and transform struct definitions
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}

		var boolFields []string
		for _, field := range st.Fields.List {
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
				for _, name := range field.Names {
					boolFields = append(boolFields, name.Name)
				}
			}
		}

		if len(boolFields) > 0 {
			originalName := ts.Name.Name
			newName := originalName + "Flags"
			packageName := file.Name.Name

			// Register in context
			ctx.AddStruct(packageName, originalName, newName, boolFields)

			fmt.Printf("    📝 Converting struct %s → %s (%d bool fields)\n",
				originalName, newName, len(boolFields))

			// Transform struct name
			ts.Name.Name = newName

			// Replace bool fields with single flags field
			st.Fields.List = []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "flags"}},
					Type:  &ast.Ident{Name: "uint64"},
				},
			}

			structsFound++
			fieldsTransformed += len(boolFields)
		}
		return true
	})

	// Second pass: Add flag constants
	if structsFound > 0 {
		p.addFlagConstants(file, ctx)
	}

	// Third pass: Transform usage patterns
	if structsFound > 0 {
		p.transformUsagePatterns(file, ctx)
	}

	if structsFound > 0 {
		fmt.Printf("    ✅ BoolToFlags: %d structs, %d fields transformed\n", structsFound, fieldsTransformed)
	}

	return nil
}

// addFlagConstants adds flag constant declarations to the file
func (p *BoolToFlagsPass) addFlagConstants(file *ast.File, ctx *TranspileContext) {
	for structName, structInfo := range ctx.Structs {
		specs := []ast.Spec{}

		for i, field := range structInfo.BoolFields {
			flagName := structInfo.FlagMapping[field]

			spec := &ast.ValueSpec{
				Names: []*ast.Ident{{Name: flagName}},
				Type:  &ast.Ident{Name: "uint64"},
				Values: []ast.Expr{
					&ast.BinaryExpr{
						X: &ast.BasicLit{
							Kind:  token.INT,
							Value: "1",
						},
						Op: token.SHL,
						Y: &ast.BasicLit{
							Kind:  token.INT,
							Value: fmt.Sprintf("%d", i),
						},
					},
				},
			}
			specs = append(specs, spec)
		}

		// Create const declaration
		constDecl := &ast.GenDecl{
			Tok:   token.CONST,
			Specs: specs,
		}

		// Insert after imports
		insertPos := 0
		for i, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				insertPos = i + 1
			}
		}

		// Insert the const declaration
		newDecls := make([]ast.Decl, 0, len(file.Decls)+1)
		newDecls = append(newDecls, file.Decls[:insertPos]...)
		newDecls = append(newDecls, constDecl)
		newDecls = append(newDecls, file.Decls[insertPos:]...)
		file.Decls = newDecls

		flagNames := make([]string, len(structInfo.BoolFields))
		for i, field := range structInfo.BoolFields {
			flagNames[i] = structInfo.FlagMapping[field]
		}
		fmt.Printf("      🏷️  Added constants for %s: %v\n", structName, flagNames)
	}
}

// transformUsagePatterns transforms struct usage patterns (assignments, conditions, etc.)
func (p *BoolToFlagsPass) transformUsagePatterns(file *ast.File, ctx *TranspileContext) {
	transformedAssignments := 0
	transformedConditions := 0

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			// Transform struct literals
			if ident, ok := node.Type.(*ast.Ident); ok {
				if structInfo, exists := ctx.Structs[ident.Name]; exists {
					p.transformCompositeLit(node, structInfo)
				}
			}
		case *ast.AssignStmt:
			// Transform assignments like cfg.Debug = true
			if p.transformAssignment(node, ctx) {
				transformedAssignments++
			}
		case *ast.IfStmt:
			// Transform if conditions like if cfg.Debug
			if p.transformIfCondition(node, ctx) {
				transformedConditions++
			}
		}
		return true
	})

	if transformedAssignments > 0 || transformedConditions > 0 {
		fmt.Printf("      ⚡ Transformed: %d assignments, %d conditions\n",
			transformedAssignments, transformedConditions)
	}
}

// transformCompositeLit transforms struct literal initialization
func (p *BoolToFlagsPass) transformCompositeLit(node *ast.CompositeLit, structInfo *StructInfo) {
	// Change type to new flags type
	node.Type = &ast.Ident{Name: structInfo.NewName}

	// Clear elements - will need statement-level transformation for proper initialization
	node.Elts = nil

	fmt.Printf("        🔄 Transformed struct literal for %s\n", structInfo.OriginalName)
}

// transformAssignment transforms assignments like cfg.Debug = true/false
func (p *BoolToFlagsPass) transformAssignment(node *ast.AssignStmt, ctx *TranspileContext) bool {
	if len(node.Lhs) != 1 || len(node.Rhs) != 1 {
		return false
	}

	sel, ok := node.Lhs[0].(*ast.SelectorExpr)
	if !ok {
		return false
	}

	val, ok := node.Rhs[0].(*ast.Ident)
	if !ok {
		return false
	}

	fieldName := sel.Sel.Name

	// Check if this is a bool field assignment using context
	for _, structInfo := range ctx.Structs {
		if ctx.IsBoolField(structInfo.OriginalName, fieldName) {
			flagName := structInfo.FlagMapping[fieldName]

			if val.Name == "true" {
				// Transform to: obj.flags |= FlagField
				node.Lhs[0] = &ast.SelectorExpr{
					X:   sel.X,
					Sel: &ast.Ident{Name: "flags"},
				}
				node.Tok = token.OR_ASSIGN
				node.Rhs[0] = &ast.Ident{Name: flagName}
			} else if val.Name == "false" {
				// Transform to: obj.flags &^= FlagField
				node.Lhs[0] = &ast.SelectorExpr{
					X:   sel.X,
					Sel: &ast.Ident{Name: "flags"},
				}
				node.Tok = token.AND_NOT_ASSIGN
				node.Rhs[0] = &ast.Ident{Name: flagName}
			}
			return true
		}
	}
	return false
}

// transformIfCondition transforms if conditions to use bitwise checks
func (p *BoolToFlagsPass) transformIfCondition(node *ast.IfStmt, ctx *TranspileContext) bool {
	// Look for conditions like "if cfg.Debug"
	if sel, ok := node.Cond.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			fieldName := sel.Sel.Name

			// Check if this is a flag field using context
			for _, structInfo := range ctx.Structs {
				if ctx.IsBoolField(structInfo.OriginalName, fieldName) {
					flagName := structInfo.FlagMapping[fieldName]

					// Transform to bitwise check: cfg.flags & FlagDebug != 0
					node.Cond = &ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: ident.Name},
								Sel: &ast.Ident{Name: "flags"},
							},
							Op: token.AND,
							Y:  &ast.Ident{Name: flagName},
						},
						Op: token.NEQ,
						Y: &ast.BasicLit{
							Kind:  token.INT,
							Value: "0",
						},
					}
					return true
				}
			}
		}
	}
	return false
}
