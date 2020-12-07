package mux

import (
	"context"
	"testing"
)

type loggerConfig struct {
	Skipper
	t *testing.T
}

func logger(config loggerConfig) Middleware {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	return Middleware(func(next Handler) Handler {
		return HandlerFunc(func(c Context) (err error) {
			if config.Skipper(c) {
				return next.Handle(c)
			}
			config.t.Logf("log: %v %+v", c.Path(), c.Args())
			return next.Handle(c)
		})
	})
}

func Test_Router(t *testing.T) {

	Use(logger(loggerConfig{t: t}))

	RegisterFunc("/users/", func(c Context) error {
		t.Logf("enter /users/: %v %+v.", c.Path(), c.Args())
		return nil
	})

	RegisterFunc("/users/list", func(c Context) error {
		t.Logf("enter /users/list: %v %+v.", c.Path(), c.Args())
		return nil
	})

	RegisterFunc("/", func(c Context) error {
		t.Logf("enter /: %v %+v.", c.Path(), c.Args())
		return nil
	})

	var err error
	err = Dispatch(context.TODO(), "/home", "hello", "world")
	if err != nil {
		t.Error(err)
	}
	err = Dispatch(context.TODO(), "/users/list/1/10", "hi", "Tom")
	if err != nil {
		t.Error(err)
	}

}
