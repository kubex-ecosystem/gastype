// Package transpiler implements logical context analysis for advanced transpilation
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// LogicalContext represents a logical context that can be transpiled
type LogicalContext struct {
	Type         string     // "function", "if_chain", "switch", "struct", "auth_logic"
	Name         string     // Context name (function, variable, etc.)
	Complexity   int        // Complexity level (1-10)
	Transpilable bool       // If it can be transpiled to bitwise
	Dependencies []string   // Dependencies on other contexts
	ASTNode      ast.Node   // Original AST node
	BitStates    []BitState // Possible bitwise states
	StartLine    int        // Start line
	EndLine      int        // End line
}

// BitState represents a possible bitwise state
type BitState struct {
	Name        string   // Semantic name of the state
	BitPosition int      // Bit position (0-63)
	Conditions  []string // Conditions that activate this state
	Actions     []string // Actions executed when active
}

// ContextAnalyzer analyzes Go code and identifies transpilable contexts
type ContextAnalyzer struct {
	fset     *token.FileSet
	contexts []LogicalContext
}

// NewContextAnalyzer creates a new context analyzer
func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{
		fset:     token.NewFileSet(),
		contexts: []LogicalContext{},
	}
}

// AnalyzeFile analyzes a file and identifies logical contexts
func (ca *ContextAnalyzer) AnalyzeFile(filename string) ([]LogicalContext, error) {
	node, err := parser.ParseFile(ca.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %w", err)
	}

	ca.contexts = []LogicalContext{}

	// Analysis of different contexts
	ca.analyzeStructs(node)
	ca.analyzeFunctions(node)
	ca.analyzeIfChains(node)
	ca.analyzeAuthLogic(node)
	ca.analyzeSwitchStatements(node)

	// Calculate transpilability and dependencies
	ca.calculateTranspilability()

	return ca.contexts, nil
}

// analyzeStructs analyzes structs for bitwise conversion (already implemented)
func (ca *ContextAnalyzer) analyzeStructs(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				context := ca.analyzeStruct(typeSpec, structType)
				if context != nil {
					ca.contexts = append(ca.contexts, *context)
				}
			}
		}
		return true
	})
}

// analyzeStruct analyzes a specific struct
func (ca *ContextAnalyzer) analyzeStruct(typeSpec *ast.TypeSpec, structType *ast.StructType) *LogicalContext {
	var boolFields []string
	var bitStates []BitState

	for _, field := range structType.Fields.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			for _, name := range field.Names {
				boolFields = append(boolFields, name.Name)
				bitStates = append(bitStates, BitState{
					Name:        name.Name,
					BitPosition: len(bitStates),
					Conditions:  []string{fmt.Sprintf("Set%s(true)", name.Name)},
					Actions:     []string{fmt.Sprintf("Execute_%s_Logic", name.Name)},
				})
			}
		}
	}

	if len(boolFields) == 0 {
		return nil
	}

	pos := ca.fset.Position(typeSpec.Pos())
	endPos := ca.fset.Position(typeSpec.End())

	return &LogicalContext{
		Type:         "struct",
		Name:         typeSpec.Name.Name,
		Complexity:   len(boolFields), // Complexity = number of bool fields
		Transpilable: true,
		Dependencies: []string{},
		ASTNode:      typeSpec,
		BitStates:    bitStates,
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}
}

// analyzeFunctions analyzes functions for obfuscation and transpilation
func (ca *ContextAnalyzer) analyzeFunctions(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			context := ca.analyzeFunction(funcDecl)
			if context != nil {
				ca.contexts = append(ca.contexts, *context)
			}
		}
		return true
	})
}

// analyzeFunction analyzes a specific function
func (ca *ContextAnalyzer) analyzeFunction(funcDecl *ast.FuncDecl) *LogicalContext {
	if funcDecl.Name == nil {
		return nil
	}

	pos := ca.fset.Position(funcDecl.Pos())
	endPos := ca.fset.Position(funcDecl.End())

	// Analyze function complexity
	complexity := ca.calculateFunctionComplexity(funcDecl)
	//transpilable := complexity <= 5 // Funções simples são transpiláveis
	transpilable := complexity <= 3 // Funções simples são transpiláveis

	context := &LogicalContext{
		Type:         "function",
		Name:         funcDecl.Name.Name,
		Complexity:   complexity,
		Transpilable: transpilable,
		Dependencies: []string{},
		ASTNode:      funcDecl,
		BitStates:    []BitState{},
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}

	// If transpilable, create bitwise states
	if transpilable {
		context.BitStates = ca.createFunctionBitStates(funcDecl)
	}

	return context
}

// calculateFunctionComplexity calculates complexity of a function
func (ca *ContextAnalyzer) calculateFunctionComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1 // Base

	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt:
			complexity += 2
		case *ast.ForStmt, *ast.RangeStmt:
			complexity += 3
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity += 2
		case *ast.CallExpr:
			complexity += 1
		}
		return true
	})

	return complexity
}

// createFunctionBitStates creates bitwise states for a function
func (ca *ContextAnalyzer) createFunctionBitStates(funcDecl *ast.FuncDecl) []BitState {
	states := []BitState{
		{
			Name:        "function_entry",
			BitPosition: 0,
			Conditions:  []string{"function_called"},
			Actions:     []string{"initialize_context"},
		},
		{
			Name:        "function_processing",
			BitPosition: 1,
			Conditions:  []string{"parameters_validated"},
			Actions:     []string{"execute_logic"},
		},
		{
			Name:        "function_exit",
			BitPosition: 2,
			Conditions:  []string{"logic_completed"},
			Actions:     []string{"return_result"},
		},
	}

	return states
}

// analyzeIfChains analyzes if/else chains for conversion to jump tables
func (ca *ContextAnalyzer) analyzeIfChains(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if ifStmt, ok := n.(*ast.IfStmt); ok {
			context := ca.analyzeIfChain(ifStmt)
			if context != nil {
				ca.contexts = append(ca.contexts, *context)
			}
		}
		return true
	})
}

// analyzeIfChain analyzes an if/else chain
func (ca *ContextAnalyzer) analyzeIfChain(ifStmt *ast.IfStmt) *LogicalContext {
	chainLength := ca.calculateIfChainLength(ifStmt)

	// TODO: Avaliar complexidade das condições
	// Somente otimizar se tiver 3 ou mais condições
	// if chainLength < 2 {
	// 	return nil // Muito simples para otimizar
	// }

	pos := ca.fset.Position(ifStmt.Pos())
	endPos := ca.fset.Position(ifStmt.End())

	context := &LogicalContext{
		Type:         "if_chain",
		Name:         fmt.Sprintf("if_chain_line_%d", pos.Line),
		Complexity:   chainLength,
		Transpilable: chainLength >= 3, // Cadeias com 3+ condições são transpiláveis
		Dependencies: []string{},
		ASTNode:      ifStmt,
		BitStates:    ca.createIfChainBitStates(ifStmt, chainLength),
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}

	return context
}

// calculateIfChainLength calculates the length of an if/else chain
func (ca *ContextAnalyzer) calculateIfChainLength(ifStmt *ast.IfStmt) int {
	length := 1
	current := ifStmt

	for current.Else != nil {
		length++
		if elseIf, ok := current.Else.(*ast.IfStmt); ok {
			current = elseIf
		} else {
			break
		}
	}

	return length
}

// createIfChainBitStates creates bitwise states for if/else chain
func (ca *ContextAnalyzer) createIfChainBitStates(ifStmt *ast.IfStmt, chainLength int) []BitState {
	states := []BitState{}

	for i := 0; i < chainLength; i++ {
		state := BitState{
			Name:        fmt.Sprintf("condition_%d", i),
			BitPosition: i,
			Conditions:  []string{fmt.Sprintf("evaluate_condition_%d", i)},
			Actions:     []string{fmt.Sprintf("execute_branch_%d", i)},
		}
		states = append(states, state)
	}

	return states
}

// analyzeAuthLogic analyzes authentication logic for obfuscation
func (ca *ContextAnalyzer) analyzeAuthLogic(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if ca.isAuthFunction(funcDecl) {
				context := ca.analyzeAuthFunction(funcDecl)
				if context != nil {
					ca.contexts = append(ca.contexts, *context)
				}
			}
		}
		return true
	})
}

// isAuthFunction checks if a function is for authentication
func (ca *ContextAnalyzer) isAuthFunction(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Name == nil {
		return false
	}

	name := strings.ToLower(funcDecl.Name.Name)
	authKeywords := []string{"auth", "login", "check", "verify", "validate", "permission"}

	for _, keyword := range authKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}

	return false
}

// analyzeAuthFunction analyzes authentication function
func (ca *ContextAnalyzer) analyzeAuthFunction(funcDecl *ast.FuncDecl) *LogicalContext {
	pos := ca.fset.Position(funcDecl.Pos())
	endPos := ca.fset.Position(funcDecl.End())

	context := &LogicalContext{
		Type:         "auth_logic",
		Name:         funcDecl.Name.Name,
		Complexity:   8,    // Autenticação é sempre complexa
		Transpilable: true, // Sempre transpilável para segurança
		Dependencies: []string{},
		ASTNode:      funcDecl,
		BitStates:    ca.createAuthBitStates(),
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}

	return context
}

// createAuthBitStates creates bitwise states for authentication
func (ca *ContextAnalyzer) createAuthBitStates() []BitState {
	return []BitState{
		{
			Name:        "auth_check_init",
			BitPosition: 0,
			Conditions:  []string{"user_input_received"},
			Actions:     []string{"validate_input_format"},
		},
		{
			Name:        "auth_processing",
			BitPosition: 1,
			Conditions:  []string{"input_valid"},
			Actions:     []string{"perform_obfuscated_check"},
		},
		{
			Name:        "auth_result",
			BitPosition: 2,
			Conditions:  []string{"check_completed"},
			Actions:     []string{"return_auth_status"},
		},
	}
}

// analyzeSwitchStatements analyzes switch statements for optimization
func (ca *ContextAnalyzer) analyzeSwitchStatements(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			context := ca.analyzeSwitchStatement(switchStmt)
			if context != nil {
				ca.contexts = append(ca.contexts, *context)
			}
		}
		return true
	})
}

// analyzeSwitchStatement analyzes a switch statement
func (ca *ContextAnalyzer) analyzeSwitchStatement(switchStmt *ast.SwitchStmt) *LogicalContext {
	caseCount := len(switchStmt.Body.List)

	// TODO: Avaliar complexidade dos cases
	// Somente otimizar se tiver 3 ou mais casos
	// if caseCount < 3 {
	// 	return nil // Muito simples
	// }

	pos := ca.fset.Position(switchStmt.Pos())
	endPos := ca.fset.Position(switchStmt.End())

	context := &LogicalContext{
		Type:         "switch",
		Name:         fmt.Sprintf("switch_line_%d", pos.Line),
		Complexity:   caseCount,
		Transpilable: true,
		Dependencies: []string{},
		ASTNode:      switchStmt,
		BitStates:    ca.createSwitchBitStates(caseCount),
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}

	return context
}

// createSwitchBitStates creates bitwise states for switch
func (ca *ContextAnalyzer) createSwitchBitStates(caseCount int) []BitState {
	states := []BitState{}

	for i := 0; i < caseCount; i++ {
		state := BitState{
			Name:        fmt.Sprintf("case_%d", i),
			BitPosition: i,
			Conditions:  []string{fmt.Sprintf("match_case_%d", i)},
			Actions:     []string{fmt.Sprintf("execute_case_%d", i)},
		}
		states = append(states, state)
	}

	return states
}

// calculateTranspilability calculates the transpilability of all contexts
func (ca *ContextAnalyzer) calculateTranspilability() {
	for i := range ca.contexts {
		context := &ca.contexts[i]

		// Factors that influence transpilability
		if context.Complexity > 10 {
			context.Transpilable = false
		}

		if context.Type == "auth_logic" {
			context.Transpilable = true // Always transpilable for security
		}

		// Calculate dependencies between contexts
		ca.calculateDependencies(context)
	}
}

// calculateDependencies calculates dependencies between contexts
func (ca *ContextAnalyzer) calculateDependencies(context *LogicalContext) {
	// TODO: Implement dependency analysis
	// SIMPLE IMPLEMENTATION TEST
	context.Dependencies = []string{}
	if context.Type == "function" && strings.Contains(strings.ToLower(context.Name), "handler") {
		// Handler functions usually depend on structs
		for _, ctx := range ca.contexts {
			if ctx.Type == "struct" {
				context.Dependencies = append(context.Dependencies, ctx.Name)
			}
		}
	}
}

// GetTranspilableContexts returns only transpilable contexts
func (ca *ContextAnalyzer) GetTranspilableContexts() []LogicalContext {
	var transpilable []LogicalContext

	for _, context := range ca.contexts {
		if context.Transpilable {
			transpilable = append(transpilable, context)
		}
	}

	return transpilable
}

// GetContextsByType returns contexts of a specific type
func (ca *ContextAnalyzer) GetContextsByType(contextType string) []LogicalContext {
	var filtered []LogicalContext

	for _, context := range ca.contexts {
		if context.Type == contextType {
			filtered = append(filtered, context)
		}
	}

	return filtered
}
