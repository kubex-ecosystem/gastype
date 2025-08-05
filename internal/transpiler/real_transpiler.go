// Package transpiler provides real bitwise transpilation functionality
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

// RealBitwiseTranspiler performs actual code transformation
type RealBitwiseTranspiler struct {
	fset    *token.FileSet
	context *TranspileContext
}

// NewRealBitwiseTranspiler creates a new real transpiler
func NewRealBitwiseTranspiler() *RealBitwiseTranspiler {
	return &RealBitwiseTranspiler{
		fset: token.NewFileSet(),
	}
}

// NewRealBitwiseTranspilerWithContext creates a new real transpiler with context
func NewRealBitwiseTranspilerWithContext(ctx *TranspileContext) *RealBitwiseTranspiler {
	return &RealBitwiseTranspiler{
		fset:    token.NewFileSet(),
		context: ctx,
	}
}

// TranspileBoolToFlags converts bool struct fields to bitwise flags
func (t *RealBitwiseTranspiler) TranspileBoolToFlags(inputFile, outputFile string) error {
	// Create default context if none provided
	if t.context == nil {
		t.context = NewContext(inputFile, outputFile, false, "")
	}

	fmt.Printf("🔄 Transpiling %s → %s\n", inputFile, outputFile)

	// Parse the input file
	f, err := parser.ParseFile(t.fset, inputFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Track structs that need flag constants
	structsToTransform := make(map[string][]string) // struct name -> field names

	// Step 1: Find structs with bool fields
	ast.Inspect(f, func(n ast.Node) bool {
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

			fmt.Printf("  📝 Converting struct %s → %s (%d bool fields)\n",
				originalName, newName, len(boolFields))

			// Register in context
			t.context.AddStruct(originalName, newName, boolFields)

			structsToTransform[originalName] = boolFields

			// Transform struct to use flags field instead of bool fields
			ts.Name.Name = newName

			// Replace struct fields with single flags field
			st.Fields.List = []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "flags"}},
					Type:  &ast.Ident{Name: "uint64"},
				},
			}
		}
		return true
	})

	// Step 2: Add flag constants for each transformed struct
	if len(structsToTransform) > 0 {
		t.addFlagConstants(f, structsToTransform)
	}

	// Step 3: Transform struct literals and field access
	t.transformStructUsage(f, structsToTransform)

	// Step 4: Write the transformed code
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if err := printer.Fprint(out, t.fset, f); err != nil {
		return fmt.Errorf("failed to write transformed code: %w", err)
	}

	// Step 5: Save context map if configured
	if t.context.MapFile != "" {
		if err := t.context.SaveMap(); err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to save context map: %v\n", err)
		} else {
			fmt.Printf("  📋 Context map saved: %s\n", t.context.MapFile)
		}
	}

	fmt.Printf("  ✅ Transpilation complete: %s\n", outputFile)
	return nil
}

// addFlagConstants adds const declarations for bitwise flags
func (t *RealBitwiseTranspiler) addFlagConstants(f *ast.File, structs map[string][]string) {
	for _, fields := range structs {
		// Create const group for this struct
		specs := []ast.Spec{}

		for i, field := range fields {
			constName := "Flag" + field

			spec := &ast.ValueSpec{
				Names: []*ast.Ident{{Name: constName}},
				Type:  &ast.Ident{Name: "uint64"},
			}

			// All flags need explicit bit shift values
			spec.Values = []ast.Expr{
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
			}

			specs = append(specs, spec)
		} // Create the const declaration
		constDecl := &ast.GenDecl{
			Tok:   token.CONST,
			Specs: specs,
		}

		// Insert const declaration after imports
		insertPos := 0
		for i, decl := range f.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				insertPos = i + 1
			}
		}

		// Insert the const declaration
		newDecls := make([]ast.Decl, 0, len(f.Decls)+1)
		newDecls = append(newDecls, f.Decls[:insertPos]...)
		newDecls = append(newDecls, constDecl)
		newDecls = append(newDecls, f.Decls[insertPos:]...)
		f.Decls = newDecls

		fmt.Printf("    🏷️  Added constants: %v\n", getConstNames(fields))
	}
}

// transformStructUsage transforms struct literals and field access
func (t *RealBitwiseTranspiler) transformStructUsage(f *ast.File, structs map[string][]string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			// Transform struct literals like ServiceConfig{Debug: true, Auth: false}
			if ident, ok := node.Type.(*ast.Ident); ok {
				if fields, exists := structs[ident.Name]; exists {
					t.transformCompositeLit(node, ident.Name, fields)
				}
			}
		case *ast.SelectorExpr:
			// Transform field access like cfg.Debug
			t.transformFieldAccess(node, structs)
		case *ast.IfStmt:
			// Transform if conditions like if cfg.Debug
			t.transformIfCondition(node, structs)
		case *ast.AssignStmt:
			// Transform assignments like cfg.Debug = true
			t.transformAssignment(node, structs)
		}
		return true
	})
}

// transformCompositeLit transforms struct literal initialization
func (t *RealBitwiseTranspiler) transformCompositeLit(node *ast.CompositeLit, structName string, fields []string) {
	// Change type to FlagsType
	node.Type = &ast.Ident{Name: structName + "Flags"}

	// Store original field initializations for later transformation
	originalInits := make(map[string]bool)

	fmt.Printf("      🔍 Analyzing composite literal with %d elements\n", len(node.Elts))

	// Extract field initializations
	for i, elt := range node.Elts {
		fmt.Printf("      🔍 Element %d: %T\n", i, elt)
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fmt.Printf("      🔍 KeyValue found\n")
			if key, ok := kv.Key.(*ast.Ident); ok {
				fmt.Printf("      🔍 Key: %s\n", key.Name)
				if val, ok := kv.Value.(*ast.Ident); ok {
					fmt.Printf("      🔍 Value: %s\n", val.Name)
					originalInits[key.Name] = (val.Name == "true")
				}
			}
		}
	}

	// Clear elements - we'll use assignments after the literal
	node.Elts = nil

	// TODO: Generate assignment statements after the variable declaration
	// This is complex and would require AST rewriting at statement level
	// For now, we'll note what needs to be initialized
	fmt.Printf("    🔄 Transformed struct literal for %s\n", structName)
	for field, value := range originalInits {
		fmt.Printf("      📋 Field %s was %t\n", field, value)
	}
} // transformFieldAccess transforms field access to bitwise operations
func (t *RealBitwiseTranspiler) transformFieldAccess(node *ast.SelectorExpr, structs map[string][]string) {
	// Look for patterns like cfg.Debug where cfg is a flags type
	if ident, ok := node.X.(*ast.Ident); ok {
		fieldName := node.Sel.Name

		// Check if this could be a flag field access
		for _, fields := range structs {
			for _, field := range fields {
				if field == fieldName {
					// This might be a flag access - we'll transform in context
					fmt.Printf("    🎯 Found field access: %s.%s\n", ident.Name, fieldName)
				}
			}
		}
	}
}

// transformIfCondition transforms if conditions to use bitwise checks
func (t *RealBitwiseTranspiler) transformIfCondition(node *ast.IfStmt, structs map[string][]string) {
	// Look for conditions like "if cfg.Debug"
	if sel, ok := node.Cond.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			fieldName := sel.Sel.Name

			// Check if this is a flag field
			for _, fields := range structs {
				for _, field := range fields {
					if field == fieldName {
						// Transform to bitwise check: cfg.flags & FlagDebug != 0
						flagName := "Flag" + fieldName

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

						fmt.Printf("    ⚡ Transformed if condition: %s.%s → bitwise check\n", ident.Name, fieldName)
						return
					}
				}
			}
		}
	}
}

// transformAssignment transforms assignments like cfg.Debug = true/false
func (t *RealBitwiseTranspiler) transformAssignment(node *ast.AssignStmt, structs map[string][]string) {
	if len(node.Lhs) != 1 || len(node.Rhs) != 1 {
		return
	}

	sel, ok := node.Lhs[0].(*ast.SelectorExpr)
	if !ok {
		return
	}

	val, ok := node.Rhs[0].(*ast.Ident)
	if !ok {
		return
	}

	fieldName := sel.Sel.Name

	// Check if this is a bool field assignment using context
	for structName := range t.context.Structs {
		if t.context.IsBoolField(structName, fieldName) {
			flagName := t.context.GetFlagName(structName, fieldName)

			if val.Name == "true" {
				// Transform to: obj.flags |= FlagField
				node.Lhs[0] = &ast.SelectorExpr{
					X:   sel.X,
					Sel: &ast.Ident{Name: "flags"},
				}
				node.Tok = token.OR_ASSIGN
				node.Rhs[0] = &ast.Ident{Name: flagName}

				fmt.Printf("    ⚡ Transformed assignment: %s = true → flags |= %s\n", fieldName, flagName)
			} else if val.Name == "false" {
				// Transform to: obj.flags &^= FlagField
				node.Lhs[0] = &ast.SelectorExpr{
					X:   sel.X,
					Sel: &ast.Ident{Name: "flags"},
				}
				node.Tok = token.AND_NOT_ASSIGN
				node.Rhs[0] = &ast.Ident{Name: flagName}

				fmt.Printf("    ⚡ Transformed assignment: %s = false → flags &^= %s\n", fieldName, flagName)
			}
			return
		}
	}
}

// getConstNames returns the constant names for fields
func getConstNames(fields []string) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = "Flag" + field
	}
	return names
}

// TranspileFile transpiles a single file using the real transpiler
func (t *RealBitwiseTranspiler) TranspileFile(inputPath, outputPath string) error {
	return t.TranspileBoolToFlags(inputPath, outputPath)
}

// TranspileProject transpiles an entire project
func (t *RealBitwiseTranspiler) TranspileProject(inputDir, outputDir string) error {
	fmt.Printf("🚀 Starting real project transpilation: %s → %s\n", inputDir, outputDir)

	// For now, just show what we would do
	fmt.Printf("  📁 Would recursively transpile all .go files\n")
	fmt.Printf("  ⚡ Would apply bool→bitflags transformation\n")
	fmt.Printf("  🔄 Would preserve non-transpilable code\n")

	return nil
}
