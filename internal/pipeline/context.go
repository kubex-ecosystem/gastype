// Package pipeline contém a definição do contexto compartilhado entre os estágios do pipeline.
package pipeline

import (
	"go/ast"
	"go/token"
	"go/types"
	"time"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
	l "github.com/rafa-mori/logz"
)

type Options struct {
	ShortStringMinLen int  // heurística (ex.: 4)
	DryRun            bool // não reescreve, só reporta
	MaxWorkers        int  // p/ Processor (tarefas assíncronas auxiliares)
}

type Stats struct {
	Visited  int64
	Changed  int64
	Skipped  int64
	ByReason map[string]int64 // "prohibited-zone","short","const-expr","idempotent","unresolved-type"
}

type Logger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
}

type Context struct {
	Fset    *token.FileSet
	Info    *types.Info
	Pkg     *types.Package
	File    *ast.File
	Opts    Options
	Metrics *Stats
	Log     gl.GLog[l.Logger]
	Bus     *EventBus
	Clock   func() time.Time
}

func NewContext(fset *token.FileSet, info *types.Info, pkg *types.Package, file *ast.File, opts Options, log gl.GLog[l.Logger]) *Context {
	return &Context{
		Fset:    fset,
		Info:    info,
		Pkg:     pkg,
		File:    file,
		Opts:    opts,
		Log:     log,
		Metrics: &Stats{ByReason: map[string]int64{}},
		Bus:     NewEventBus(),
		Clock:   time.Now,
	}
}
