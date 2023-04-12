package logger

import (
	"errors"
	"github.com/billyyoyo/microj/errs"
	"testing"
)

func TestLogger(t *testing.T) {
	Debug("hello world")
	Info("hello world")
	Warn("hello world")
	Error("fuck", test1(), Val{"", ""})
	Fatal("broken", test1())
}

func TestError(t *testing.T) {
	Error("fuck", test1())
}

func test1() error {
	return test2()
}

func test2() error {
	return errs.Wrap(500510, "bad db", test3())
}

func test3() error {
	return errors.New("db can not access")
}
