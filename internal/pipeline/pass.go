package pipeline

import "go/ast"

type Constraint struct {
	Deterministic bool
	Idempotent    bool
	Reentrant     bool
}

type Pass interface {
	Name() string
	Priority() int
	Constraints() Constraint
	Apply(file *ast.File, ctx *Context) error
}
