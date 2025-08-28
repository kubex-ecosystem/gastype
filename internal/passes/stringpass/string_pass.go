// Package stringpass contém passes de transformação para strings.
package stringpass

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/rafa-mori/gastype/internal/pipeline"
	"github.com/rafa-mori/gastype/internal/xast"
	"github.com/rafa-mori/gastype/internal/xtype"
	"golang.org/x/tools/go/ast/astutil"
)

type Pass struct{}

func New() *Pass { return &Pass{} }

func (p *Pass) Name() string  { return "StringObfuscate" }
func (p *Pass) Priority() int { return 100 }
func (p *Pass) Constraints() pipeline.Constraint {
	return pipeline.Constraint{Deterministic: true, Idempotent: true, Reentrant: true}
}

func (p *Pass) Apply(file *ast.File, ctx *pipeline.Context) error {
	pre := func(c *astutil.Cursor) bool {
		n := c.Node()
		bl, ok := n.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return true
		}

		// zonas proibidas
		if xast.InImportOrTag((*xast.Cursor)(c)) {
			ctx.Metrics.Skipped++
			ctx.Metrics.ByReason["prohibited-zone"]++
			return true
		}

		raw := bl.Value
		if len(raw) < 2 {
			return true
		}
		val, err := strconv.Unquote(raw)
		if err != nil {
			ctx.Metrics.Skipped++
			ctx.Metrics.ByReason["unquote-error"]++
			return true
		}
		if len(val) < ctx.Opts.ShortStringMinLen {
			ctx.Metrics.Skipped++
			ctx.Metrics.ByReason["short"]++
			return true
		}

		// const expr -> skip (v1)
		if xtype.IsConstString(ctx.Info, bl) {
			ctx.Metrics.Skipped++
			ctx.Metrics.ByReason["const-expr"]++
			return true
		}

		// já transformado? idempotência
		if xast.IsObfStringCall(n) {
			ctx.Metrics.Skipped++
			ctx.Metrics.ByReason["idempotent"]++
			return true
		}

		// tipo nominal local (preserva LogType/pkg.LogType quando existir)
		typeExpr := xast.ResolveLocalStringType((*xast.Cursor)(c), ctx.Info)
		newNode := xast.MakeStringFromBytes(typeExpr, []byte(val))

		if ctx.Opts.DryRun {
			ctx.Metrics.Visited++
			return true
		}

		c.Replace(newNode)
		ctx.Metrics.Changed++
		ctx.Bus.Publish(pipeline.Event{Topic: pipeline.TopicNodeChanged, Payload: ctx.Fset.Position(n.Pos())})

		return true
	}

	post := func(*astutil.Cursor) bool { return true }

	return xast.NewWalker().Walk(file, pre, post)
}
