package errors

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("err")
	fmt.Println(err)
}

func TestWarp(t *testing.T) {
	err := New("err1")
	err = Wrap(err)
	err = Wrap(err)
	err = Wrap(err)
	err = Wrap(err)
	fmt.Println(err)
}

func TestWarpWithMsg(t *testing.T) {
	err := New("err1")
	err = Wrap(err)
	err = Wrap(err)
	err = Wrap(err)
	err = WrapWithMsg(err, "last")
	fmt.Println(err)
}
