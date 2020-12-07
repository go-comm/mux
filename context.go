package mux

import "context"

func BorrowContext(ctx context.Context, path string, args ...interface{}) Context {
	return defaultServeMux.BorrowContext(ctx, path, args...)
}

func ReturnContext(c Context) {
	defaultServeMux.ReturnContext(c)
}

func (m *ServeMux) BorrowContext(ctx context.Context, path string, args ...interface{}) Context {
	return &baseContext{ctx: ctx, mux: m, path: path, args: args}
}

func (m *ServeMux) ReturnContext(c Context) {
	if bc, ok := c.(*baseContext); ok {
		bc.ctx = nil
		bc.mux = nil
		bc.path = ""
		bc.args = nil
	}
}

type Context interface {
	Context() context.Context
	Path() string
	Args() []interface{}
	ServeMux() *ServeMux
}

var _ Context = (*baseContext)(nil)

type baseContext struct {
	mux  *ServeMux
	ctx  context.Context
	path string
	args []interface{}
}

func (c *baseContext) Context() context.Context {
	return c.ctx
}

func (c *baseContext) Path() string {
	return c.path
}

func (c *baseContext) Args() []interface{} {
	return c.args
}

func (c *baseContext) ServeMux() *ServeMux {
	return c.mux
}
