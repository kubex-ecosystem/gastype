package xast

import (
	"go/ast"

	"golang.org/x/tools/go/ast/astutil"
)

type Cursor = astutil.Cursor

type Walker interface {
	Walk(file *ast.File, pre func(*Cursor) bool, post func(*Cursor) bool) error
}

type applyWalker struct{}

func NewWalker() Walker { return &applyWalker{} }

func (w *applyWalker) Walk(file *ast.File, pre func(*Cursor) bool, post func(*Cursor) bool) error {
	astutil.Apply(file,
		func(c *astutil.Cursor) bool { return pre(c) },
		func(c *astutil.Cursor) bool { return post(c) },
	)
	return nil
}
