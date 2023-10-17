package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
)

func TestAs(t *testing.T) {
	e := New(100, "afdf")
	var me MicroError
	b := errors.As(e, &me)
	fmt.Printf("ret %v  |  %s\n", b, me.Error())

}
