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
	ERRCODE_REQ_PARAMETER
	ERRCODE_NO_TOKEN
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type IsMicro interface {
	IsMicro()
}

type MicroError struct {
	code int
	msg  string
	err  error
}

func (e *MicroError) As(tar any) bool {
	er, ok := tar.(*MicroError)
	if ok {
		er.code = e.code
		er.msg = e.msg
		er.err = e.err
	}
	return ok
}

func (e MicroError) IsMicro() {}

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
	e := &MicroError{
		code: code,
		msg:  msg,
		err:  errors.New(msg),
	}
	return e
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

func WrapRpcError(e error) error {
	if e == nil {
		return nil
	}

	if me, ok := e.(*MicroError); ok {
		return NewRpcError(me.code, me.msg)
	} else {
		return NewRpcError(ERRCODE_COMMON, e.Error())
	}
}
