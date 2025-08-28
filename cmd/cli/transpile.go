// Package cli provides pipeline commands for GASType Premium Pipeline
package cli

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"

	"github.com/rafa-mori/gastype/internal/passes/stringpass"
	"github.com/rafa-mori/gastype/internal/pipeline"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
	l "github.com/rafa-mori/logz"
)

type InputKind string

const (
	KindFileGo  InputKind = "file.go"
	KindDir     InputKind = "dir"
	KindModule  InputKind = "module"  // dir com go.mod
	KindPattern InputKind = "pattern" // ./..., std, github.com/x/...
	KindImport  InputKind = "import"  // import path simples
)

type LoadPlan struct {
	Kind    InputKind
	Dir     string   // quando aplicável (módulo/dir)
	Pattern []string // o que vai pro Load(...)
	Note    string   // dica pro usuário
}

// TranspileCmd creates the main transpile command
func TranspileCmd() *cobra.Command {
	var cfg = &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports,
	}
	var passes, patterns []string
	var outputPath, outputFormat, mapFile, mode string
	var preserveDocs, backupOriginal, noObfuscate, dryRun, estimatePerf, verbose bool
	var securityLevel int

	cmd := &cobra.Command{
		Use:   "transpile",
		Short: "Transpile Go code to bitwise-optimized equivalent",
		Long: `Transpile traditional Go code to bitwise-optimized equivalent using AST analysis.

		This command analyzes Go source code and identifies optimization opportunities:
		- Boolean struct fields → Bitwise flags
		- If/else chains → Jump tables
		- String literals → Byte arrays (security)
		- Configuration structs → Flag systems

		The transpiler can operate in different modes:
		- analyze: Only analyze and report optimization opportunities
		- transpile: Generate transpiled code files
		- both: Analyze and generate transpiled code
		- full-project: Complete project transpilation with build system
		- staged-transpile: Multi-stage transpilation (clean → validate → obfuscate)

		Examples:
		  gastype transpile -i ./src -o ./src_optimized`,

		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			initMessage := fmt.Sprintf("Starting transpilation process: input=%v", patterns)
			gl.Log("info", initMessage)

			// Load packages
			pkgs, err := packages.Load(cfg, patterns...)
			if err != nil || packages.PrintErrors(pkgs) > 0 {
				gl.Log("Fatal", err)
			}

			// Initialize Passes slice with defaults
			reg := &pipeline.Registry{}
			reg.Register(stringpass.New())

			for _, pkg := range pkgs {
				for i, f := range pkg.Syntax {
					// Get file set and type information
					fset := pkg.Fset

					// Get type information
					info := pkg.TypesInfo

					// Create pipeline context
					ctx := pipeline.NewContext(fset, info, pkg.Types, f, pipeline.Options{
						ShortStringMinLen: 4,
						DryRun:            false,
						MaxWorkers:        0,
					}, gl.GetLogger[l.Logger](nil))

					// Run the pipeline
					if err := reg.Run(f, ctx); err != nil {
						pos := fset.Position(f.Pos())
						fmt.Fprintf(os.Stderr, "file=%s index=%d error=%v\n", pos.Filename, i, err)
					}
				}
			}
		},
	}

	// Input/Output flags
	cmd.Flags().StringSliceVarP(&patterns, "input", "i", []string{"."}, "Input directory or file containing Go code to analyze/transpile")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "./gastype_output", "Output directory for transpiled code and analysis results")

	// Mode and format flags
	cmd.Flags().StringVarP(&mode, "mode", "m", "analyze", "Operation mode: analyze, transpile, both, full-project, staged-transpile")
	cmd.Flags().StringVar(&outputFormat, "format", "json", "Output format for analysis results: json, yaml, or text")

	// Optimization flags
	cmd.Flags().IntVar(&securityLevel, "security", 2, "Security optimization level (1=low, 2=medium, 3=high)")
	cmd.Flags().BoolVar(&preserveDocs, "preserve-docs", true, "Preserve original comments and documentation in transpiled code")
	cmd.Flags().BoolVar(&backupOriginal, "backup", true, "Create backup of original files before transpilation")

	// Pipeline flags
	cmd.Flags().BoolVar(&noObfuscate, "no-obfuscate", false, "Stage 1: Transpile without obfuscation (readable optimized code)")
	cmd.Flags().StringVar(&mapFile, "map", "", "Generate context mapping JSON file for transpilation tracking")

	// Engine flags
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run transpilation in dry-run mode without modifying files")
	cmd.Flags().BoolVar(&estimatePerf, "estimate-perf", false, "Estimate performance impact of transpilation")
	cmd.Flags().StringSliceVar(&passes, "passes", passes, "Specify transpilation passes to run (comma-separated)")

	// Utility flags
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed logs of analysis and transpilation process")

	return cmd
}

func PlanInputs(args []string) ([]LoadPlan, error) {
	var plans []LoadPlan
	for _, in := range args {
		if fi, err := os.Stat(in); err == nil {
			if fi.IsDir() {
				// módulo aninhado?
				if _, e := os.Stat(filepath.Join(in, "go.mod")); e == nil {
					plans = append(plans, LoadPlan{
						Kind: KindModule, Dir: in, Pattern: []string{"./..."},
						Note: "módulo detectado; usando Dir=" + in + " e padrão ./...",
					})
				} else {
					// dir “solto”: tenta dir e fallback dir/...
					p := in
					if !strings.HasPrefix(p, "./") && !strings.HasPrefix(p, "../") && !filepath.IsAbs(p) {
						p = "./" + p
					}
					plans = append(plans, LoadPlan{
						Kind: KindDir, Dir: "", Pattern: []string{p, filepath.Join(in, "...")},
						Note: "diretório detectado; tentando diretório e varredura recursiva",
					})
				}
				continue
			}
			if strings.HasSuffix(in, ".go") {
				plans = append(plans, LoadPlan{
					Kind: KindFileGo, Pattern: []string{in},
					Note: "arquivo .go detectado; carregando pacote do arquivo",
				})
				continue
			}
		}
		// fallback: pode ser import path ou pattern
		plans = append(plans, LoadPlan{
			Kind: KindPattern, Pattern: []string{in},
			Note: "tratando como pattern/import path (ex.: std, ./..., github.com/x/y/...)",
		})
	}
	return plans, nil
}

func LoadWithPlans(plans []LoadPlan, tests bool, mode packages.LoadMode, logf func(string, ...any)) ([]*packages.Package, error) {
	var acc []*packages.Package
	for _, pl := range plans {
		cfg := &packages.Config{Mode: mode, Tests: tests}
		if pl.Dir != "" {
			cfg.Dir = pl.Dir
		}

		// tente cada pattern até um dar certo
		var ok bool
		for i, pat := range pl.Pattern {
			logf("[INFO] loading kind=%s dir=%q pattern=%q (%s)", pl.Kind, cfg.Dir, pat, pl.Note)
			pkgs, err := packages.Load(cfg, pat)
			if err == nil && packages.PrintErrors(pkgs) == 0 && len(pkgs) > 0 {
				acc = append(acc, pkgs...)
				ok = true
				break
			}
			if i == len(pl.Pattern)-1 {
				// última tentativa: devolve o primeiro erro legível
				return nil, fmt.Errorf(buildNiceLoadError(err, pkgs, cfg, pat, pl))
			}
		}

		if !ok {

		}
	}
	return acc, nil
}

func buildNiceLoadError(err error, pkgs []*packages.Package, cfg *packages.Config, pat string, pl LoadPlan) string {
	var sb strings.Builder
	sb.WriteString("falha ao carregar código\n")
	sb.WriteString(fmt.Sprintf("- dir base: %q\n- pattern: %q\n- kind: %s\n", cfg.Dir, pat, pl.Kind))
	if err != nil {
		sb.WriteString(fmt.Sprintf("- erro: %v\n", err))
	}
	for _, p := range pkgs {
		for _, e := range p.Errors {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", e.Pos, e.Msg))
		}
	}
	sb.WriteString("dicas:\n")
	switch pl.Kind {
	case KindDir:
		sb.WriteString("• tente usar ./dir ou ./dir/... (não passe só o nome do dir)\n")
	case KindImport, KindPattern:
		sb.WriteString("• se for diretório local, prefixe com ./\n")
		sb.WriteString("• se for módulo aninhado, rode com -i ./modulo e garanta que existe go.mod lá\n")
	case KindFileGo:
		sb.WriteString("• confirme build tags/GOOS/GOARCH se o arquivo não entra no build\n")
	}
	return sb.String()
}

func DebugPkgSummary(p *packages.Package) string {
	files := len(p.Syntax)
	uses := 0
	for _, f := range p.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel != nil {
				if p.TypesInfo != nil && p.TypesInfo.Uses[sel.Sel] != nil {
					uses++
				}
			}
			return true
		})
	}
	modDir := ""
	if p.Module != nil {
		modDir = p.Module.Dir
	}
	return fmt.Sprintf("pkg=%s files=%d module=%q uses(sel)=%d", p.PkgPath, files, modDir, uses)
}
