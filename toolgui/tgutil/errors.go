package tgutil

import (
	"errors"
	"fmt"
	"runtime"
)

// Errorf return Error with the function info.
func Errorf(format string, args ...interface{}) error {
	pc, _, _, ok := runtime.Caller(1)
	prefix := "unknow: "
	if ok {
		prefix = fmt.Sprintf("%s: ", runtime.FuncForPC(pc).Name())
	}
	return fmt.Errorf(prefix+format, args...)
}

// NewError return Error with the function info.
func NewError(info string) error {
	pc, _, _, ok := runtime.Caller(1)
	prefix := "unknow: "
	if ok {
		prefix = fmt.Sprintf("%s: ", runtime.FuncForPC(pc).Name())
	}
	return errors.New(prefix + info)
}
