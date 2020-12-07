package mux

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	ErrPathMismatch = errors.New("mux: path mismatch")

	defaultServeMux = new(ServeMux)
	DefaultServeMux = defaultServeMux
)

type Handler interface {
	Handle(Context) error
}

type HandlerFunc func(c Context) error

func (h HandlerFunc) Handle(c Context) error {
	return h(c)
}

type wrappedHandler struct {
	pattern string
	h       Handler
}

type ServeMux struct {
	mutex       sync.RWMutex
	middlewares []Middleware
	dict        map[string]*wrappedHandler
	hs          []*wrappedHandler
}

func (m *ServeMux) Use(middlewares ...Middleware) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *ServeMux) handler(c Context) (h Handler, have bool) {
	m.mutex.RLock()
	if v, ok := m.dict[c.Path()]; ok {
		m.mutex.RUnlock()
		h = v.h
		have = true
		return
	}
	for i := len(m.hs) - 1; i >= 0; i-- {
		v := m.hs[i]
		if strings.HasPrefix(c.Path(), v.pattern) {
			h = v.h
			have = true
			break
		}
	}
	m.mutex.RUnlock()
	return
}

func (m *ServeMux) Register(pattern string, h Handler, middlewares ...Middleware) {
	if pattern == "" {
		panic(errors.New("mux: invalid pattern"))
	}
	if h == nil {
		panic(errors.New("mux: nil handler"))
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.dict == nil {
		m.dict = make(map[string]*wrappedHandler)
	}
	if _, ok := m.dict[pattern]; ok {
		panic(fmt.Errorf("mux: pattern %s has already registered", pattern))
	}
	for i := 0; i < len(m.middlewares); i++ {
		h = m.middlewares[i](h)
	}
	for i := 0; i < len(middlewares); i++ {
		h = middlewares[i](h)
	}
	m.dict[pattern] = &wrappedHandler{pattern, h}
	if pattern[len(pattern)-1] == '/' {
		m.hs = appendSorted(m.hs, &wrappedHandler{pattern, h})
	}
}

func appendSorted(s []*wrappedHandler, e *wrappedHandler) []*wrappedHandler {
	i := sort.Search(len(s), func(i int) bool {
		return len(s[i].pattern) >= len(e.pattern)
	})
	if i == len(s) {
		return append(s, e)
	}
	s = append(s, nil)
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

func deleteSorted(s []*wrappedHandler, pattern string) []*wrappedHandler {
	var p int = -1
	for i, v := range s {
		if pattern == v.pattern {
			p = i
			break
		}
	}
	if p >= 0 {
		return append(s[:p], s[p+1:]...)
	}
	return s
}

func (m *ServeMux) RegisterFunc(pattern string, h func(c Context) error, middlewares ...Middleware) {
	m.Register(pattern, HandlerFunc(h), middlewares...)
}

func (m *ServeMux) Unregister(pattern string) {
	m.mutex.Lock()
	if _, ok := m.dict[pattern]; ok {
		delete(m.dict, pattern)
		if pattern[len(pattern)-1] == '/' {
			m.hs = deleteSorted(m.hs, pattern)
		}
	}
	m.mutex.Unlock()
}

func (m *ServeMux) Handle(c Context) error {
	h, ok := m.handler(c)
	if !ok {
		return ErrPathMismatch
	}
	return h.Handle(c)
}

func (m *ServeMux) Dispatch(ctx context.Context, path string, args ...interface{}) error {
	c := m.BorrowContext(ctx, path, args...)
	err := defaultServeMux.Handle(c)
	m.ReturnContext(c)
	return err
}

func Use(middlewares ...Middleware) {
	defaultServeMux.Use(middlewares...)
}

func Register(pattern string, h Handler, middlewares ...Middleware) {
	defaultServeMux.Register(pattern, h, middlewares...)
}

func RegisterFunc(pattern string, h func(c Context) error, middlewares ...Middleware) {
	defaultServeMux.RegisterFunc(pattern, h, middlewares...)
}

func Unregister(pattern string) {
	defaultServeMux.Unregister(pattern)
}

func Handle(c Context) error {
	return defaultServeMux.Handle(c)
}

func Dispatch(ctx context.Context, path string, args ...interface{}) error {
	return defaultServeMux.Dispatch(ctx, path, args...)
}
