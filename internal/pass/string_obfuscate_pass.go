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

// NewStringObfuscatePass cria uma nova instância do StringObfuscatePass
func NewStringObfuscatePass() *StringObfuscatePass { return &StringObfuscatePass{} }

// Name retorna o nome do pass
func (p *StringObfuscatePass) Name() string { return "StringObfuscate" }

// Apply percorre a AST do arquivo e obfusca strings literais em variáveis e constantes, convertendo-as em arrays de bytes
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

	// Leitura de dados de estruturas/objetos/constantes/expressões para atuação do pass
	objectsSpecMap := p.ScanScope(file, nil)

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
		// Aqui nós já estamos capturando quase tudo de dados do nó
		// que precisamos para decidir se obfuscamos ou não, e como obfuscamos, etc...
		bl, val, isConstCheck, ok := p.FilterExpr(n, importSpecs, structTags, objectsSpecMap, ctx)
		if !ok {
			return true
		}

		// Converte cada caractere da string em seu valor byte
		// Apesar de não ter tanta diferença, usa byte ao invés de rune e alguns outros "prós"
		// Não seria necessário o escopo desse nó e o que ele trás aqui onde está, porém para criação da lógica
		// do init e pra futuras melhorias, mantém-se nesse padrão e lugar, pode e provavelmente será usado em etapas anteriores à
		// essa aqui.. hehe
		byteVals := make([]string, len(val))
		for i, b := range []byte(val) {
			byteVals[i] = strconv.Itoa(int(b))
		}

		if !objectsSpecMap[bl].isConst && !isConstCheck {
			if !astutil.CheckConstant(
				bl,
				bl.Kind.String(),
				ctx,
			) {
				// Definição mais precisa de tipo é inferida na atribuição,
				// tanto em função de var sem tipo quanto em função
				// de var com tipo ou labeled type.
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

				// Marca que houve transformação (não constantes)
				transformations++
				return true
			}

			// Se for constante, não obfusca, mas marca que houve transformação pra contabilizar spec percorrido
			// (marcação não está sendo utilizada pra nada, só logging e olhe lá.. rsrs - por enquanto!)
			transformations++
			return true
		}

		// Ignorados
		return false
	})

	// Se houver transformação, log
	if transformations > 0 {
		gl.Log("info", fmt.Sprintf("🔄 StringObfuscatePass: %d transformations applied", transformations))
	}

	return nil
}

// ScanScope é responsável por preencher/trazer o mapa de especificações dos objetos (structs, const, var, etc.)
func (p *StringObfuscatePass) ScanScope(file *ast.File, _ *token.FileSet) map[any]objectSpecMap {
	specMap := make(map[any]objectSpecMap)
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
					specMap[bl] = objectSpecMap{
						isConst:         gd.Tok == token.CONST,
						name:            name,
						varType:         varType,
						obfuscatedName:  fmt.Sprintf("obfuscated_%s", name),
						obfuscatedValue: nil,
					}
					return false
				})
			} else {
				specMap[vs.Names[0]] = objectSpecMap{
					isConst:         gd.Tok == token.CONST,
					name:            name,
					varType:         varType,
					obfuscatedName:  fmt.Sprintf("obfuscated_%s", name),
					obfuscatedValue: nil,
				}
			}
		}
	}
	return specMap
}

// FilterExpr é responsável por determinar se um nó AST é um literal string válido para obfuscação adquirir:
// 1: BasicLit (literal básico)
// 2: valor atribuído pro objeto,
// 3: bool se token atribuído é CONST de fato, atribui tipo primário
// 4: bool se valor é de fato um const, mesmo atrás de labeled, com ou sem valor necessário para definição
// Os tipos 3 e 4, apesar de praticamente idênticos, servem para diferenciar etapas de conversão de retorno para original, entre outras coisas...
// O AST não trás essa informação de forma explícita, então é necessário inferir a partir do contexto
func (p *StringObfuscatePass) FilterExpr(n ast.Node, importSpecs, structTags map[*ast.BasicLit]bool, objectsSpecMap map[any]objectSpecMap, ctx *astutil.TranspileContext) (*ast.BasicLit, string, bool, bool) {
	bl, ok := n.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return nil, "", false, false
	}

	if importSpecs[bl] || structTags[bl] {
		return nil, "", false, false
	}

	if len(bl.Value) < 2 {
		return nil, "", false, false
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
			return nil, "", false, false
		}
	}

	isConstCheck := astutil.CheckConstant(
		bl,
		bl.Kind.String(),
		ctx,
	)

	if isConstCheck {
		// Se for constante de qualquer tipo, não obfusca
		// Aqui tornaríamos elas vars, com os tipos adequados mapeados, como labeled
		// para que possam ser utilizadas em outros contextos sem restrições e problemas
		return nil, "", false, true
	}

	// Valor da strings
	val, err := strconv.Unquote(bl.Value)
	if err != nil || len(val) < 4 {
		// Se valor de string, mesmo que literal, for curto demais, ignorar
		return nil, "", false, false
	}

	for _, w := range []string{"main", "func", "package", "import", "var", "const", "if", "else", "for", "range"} {
		if val == w {
			// Palavras reservadas do Go não são obfuscadas
			return nil, "", false, false
		}
	}

	return bl, val, isConstCheck, true
}
