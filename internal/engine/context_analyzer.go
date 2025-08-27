// Package transpiler implementa análise de contextos lógicos para transpilação avançada
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// LogicalContext representa um contexto lógico que pode ser transpilado
type LogicalContext struct {
	Type         string     // "function", "if_chain", "switch", "struct", "auth_logic"
	Name         string     // Nome do contexto (função, variável, etc.)
	Complexity   int        // Nível de complexidade (1-10)
	Transpilable bool       // Se pode ser transpilado para bitwise
	Dependencies []string   // Dependências de outros contextos
	ASTNode      ast.Node   // Nó AST original
	BitStates    []BitState // Estados bitwise possíveis
	StartLine    int        // Linha de início
	EndLine      int        // Linha de fim
}

// BitState representa um estado bitwise possível
type BitState struct {
	Name        string   // Nome semântico do estado
	BitPosition int      // Posição do bit (0-63)
	Conditions  []string // Condições que ativam este estado
	Actions     []string // Ações executadas quando ativo
}

// ContextAnalyzer analisa código Go e identifica contextos transpiláveis
type ContextAnalyzer struct {
	fset     *token.FileSet
	contexts []LogicalContext
}

// NewContextAnalyzer cria um novo analisador de contextos
func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{
		fset:     token.NewFileSet(),
		contexts: []LogicalContext{},
	}
}

// AnalyzeFile analisa um arquivo e identifica contextos lógicos
func (ca *ContextAnalyzer) AnalyzeFile(filename string) ([]LogicalContext, error) {
	node, err := parser.ParseFile(ca.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear arquivo: %w", err)
	}

	ca.contexts = []LogicalContext{}

	// Análise de contextos diferentes
	ca.analyzeStructs(node)
	ca.analyzeFunctions(node)
	ca.analyzeIfChains(node)
	ca.analyzeAuthLogic(node)
	ca.analyzeSwitchStatements(node)

	// Calcular transpilabilidade e dependências
	ca.calculateTranspilability()

	return ca.contexts, nil
}

// analyzeStructs analisa structs para conversão bitwise (já implementado)
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

// analyzeStruct analisa uma struct específica
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
		Complexity:   len(boolFields), // Complexidade = número de campos bool
		Transpilable: true,
		Dependencies: []string{},
		ASTNode:      typeSpec,
		BitStates:    bitStates,
		StartLine:    pos.Line,
		EndLine:      endPos.Line,
	}
}

// analyzeFunctions analisa funções para obfuscação e transpilação
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

// analyzeFunction analisa uma função específica
func (ca *ContextAnalyzer) analyzeFunction(funcDecl *ast.FuncDecl) *LogicalContext {
	if funcDecl.Name == nil {
		return nil
	}

	pos := ca.fset.Position(funcDecl.Pos())
	endPos := ca.fset.Position(funcDecl.End())

	// Analisar complexidade da função
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

	// Se for transpilável, criar estados bitwise
	if transpilable {
		context.BitStates = ca.createFunctionBitStates(funcDecl)
	}

	return context
}

// calculateFunctionComplexity calcula complexidade de uma função
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

// createFunctionBitStates cria estados bitwise para uma função
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

// analyzeIfChains analisa cadeias de if/else para conversão em jump tables
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

// analyzeIfChain analisa uma cadeia de if/else
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

// calculateIfChainLength calcula o comprimento de uma cadeia if/else
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

// createIfChainBitStates cria estados bitwise para cadeia if/else
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

// analyzeAuthLogic analisa lógicas de autenticação para obfuscação
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

// isAuthFunction verifica se uma função é de autenticação
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

// analyzeAuthFunction analisa função de autenticação
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

// createAuthBitStates cria estados bitwise para autenticação
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

// analyzeSwitchStatements analisa switch statements para otimização
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

// analyzeSwitchStatement analisa um switch statement
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

// createSwitchBitStates cria estados bitwise para switch
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

// calculateTranspilability calcula a transpilabilidade de todos os contextos
func (ca *ContextAnalyzer) calculateTranspilability() {
	for i := range ca.contexts {
		context := &ca.contexts[i]

		// Fatores que influenciam transpilabilidade
		if context.Complexity > 10 {
			context.Transpilable = false
		}

		if context.Type == "auth_logic" {
			context.Transpilable = true // Sempre transpilável para segurança
		}

		// Calcular dependências entre contextos
		ca.calculateDependencies(context)
	}
}

// calculateDependencies calcula dependências entre contextos
func (ca *ContextAnalyzer) calculateDependencies(context *LogicalContext) {
	// TODO: Implementar análise de dependências
	// TESTE DE IMPLEMENTAÇÃO SIMPLES
	context.Dependencies = []string{}
	if context.Type == "function" && strings.Contains(strings.ToLower(context.Name), "handler") {
		// Funções handler geralmente dependem de structs
		for _, ctx := range ca.contexts {
			if ctx.Type == "struct" {
				context.Dependencies = append(context.Dependencies, ctx.Name)
			}
		}
	}
}

// GetTranspilableContexts retorna apenas contextos transpiláveis
func (ca *ContextAnalyzer) GetTranspilableContexts() []LogicalContext {
	var transpilable []LogicalContext

	for _, context := range ca.contexts {
		if context.Transpilable {
			transpilable = append(transpilable, context)
		}
	}

	return transpilable
}

// GetContextsByType retorna contextos de um tipo específico
func (ca *ContextAnalyzer) GetContextsByType(contextType string) []LogicalContext {
	var filtered []LogicalContext

	for _, context := range ca.contexts {
		if context.Type == contextType {
			filtered = append(filtered, context)
		}
	}

	return filtered
}
