package errors

import (
	oe "errors"
	"fmt"
	"runtime"
)

func New(text string) error {
	_, file, line, _ := runtime.Caller(1)
	return oe.New(fmt.Sprintf("%s at %s:%d", text, file, line))
}

func Wrap(err error) error {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("at %s:%d\n%w", file, line, err)
}

func WrapWithMsg(err error, msg string) error {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("%s at %s:%d\n%w", msg, file, line, err)
}

func Unwrap(err error) error {
	return oe.Unwrap(err)
}

func Is(err, target error) bool {
	return oe.Is(err, target)
}

func As(err error, target any) bool {
	return oe.As(err, target)
}
