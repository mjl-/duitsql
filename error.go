package main

import (
	"fmt"
)

type localError struct {
	err error
}

func errorHandler(fn func(xerr error)) (func(error, string), func()) {
	check := func(err error, msg string) {
		if err != nil {
			panic(&localError{fmt.Errorf("%s: %s", msg, err)})
		}
	}
	handle := func() {
		e := recover()
		if e == nil {
			return
		}
		if le, ok := e.(*localError); ok {
			fn(le.err)
		} else {
			panic(e)
		}
	}
	return check, handle
}
