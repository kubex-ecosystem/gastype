package pipeline

import (
	"fmt"
	"go/ast"
	"sort"
)

type Registry struct{ ps []Pass }

func (r *Registry) Register(p Pass) { r.ps = append(r.ps, p) }

func (r *Registry) Run(file *ast.File, ctx *Context) error {
	sort.SliceStable(r.ps, func(i, j int) bool { return r.ps[i].Priority() < r.ps[j].Priority() })
	for _, p := range r.ps {
		if err := p.Apply(file, ctx); err != nil {
			return fmt.Errorf("%s: %w", p.Name(), err)
		}
	}
	return nil
}
