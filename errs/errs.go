package errs

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ERRMSG_UNKNOWN = "unknown error"

const (
	ERRCODE_COMMON = iota + 500
	ERRCODE_REMOTE_CALL
	ERRCODE_BROKER
	ERRCODE_REGISTRY
	ERRCODE_CONFIG
	ERRCODE_GATEWAY

	ERRCODE_NO_TOKEN
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type MicroError struct {
	code int
	msg  string
	err  error
}

func (e MicroError) Error() string {
	return e.msg
}

func (e MicroError) Code() int {
	return e.code
}

func (e MicroError) StackTrace() errors.StackTrace {
	if st, ok := e.err.(stackTracer); ok {
		return st.StackTrace()
	} else {
		return make([]errors.Frame, 0)
	}
}

func New(code int, msg string) error {
	return &MicroError{
		code: code,
		msg:  msg,
		err:  errors.New(msg),
	}
}
func NewInternal(msg string) error {
	return New(ERRCODE_COMMON, msg)
}

func WrapInternal(msg string, err error) error {
	return Wrap(ERRCODE_COMMON, msg, err)
}

func Wrap(code int, msg string, err error) error {
	if me, ok := err.(MicroError); !ok {
		e := errors.Wrap(err, msg)
		me = MicroError{
			code: code,
			msg:  msg,
			err:  e,
		}
		return me
	} else {
		me.code = code
		me.msg = msg
		return me
	}
}

func NewRpcError(code int, msg string) error {
	err := status.New(codes.Code(code), msg).Err()
	return err
}
