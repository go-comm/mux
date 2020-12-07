package mux

import (
	"strings"
)

type Middleware func(next Handler) Handler

type Skipper func(c Context) bool

func DefaultSkipper(c Context) bool {
	return false
}

func StripPrefix(prefix string) Middleware {
	return Middleware(func(next Handler) Handler {
		return HandlerFunc(func(c Context) (err error) {
			if p := strings.TrimPrefix(c.Path(), prefix); len(p) < len(c.Path()) {
				newc := c.ServeMux().BorrowContext(c.Context(), p, c.Args()...)
				err = next.Handle(newc)
				c.ServeMux().ReturnContext(newc)
			} else {
				err = next.Handle(c)
			}
			return
		})
	})
}
