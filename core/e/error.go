package e

import "fmt"

type WrapError interface {
	Code() int
	Message() string
	Details() []string
	Error() string
}

type wrapErrorImpl struct {
	code    int
	msg     string
	details []string
}

func (err *wrapErrorImpl) Code() int {
	return err.code
}

func (err *wrapErrorImpl) Message() string {
	return err.msg
}
func (err *wrapErrorImpl) Details() []string {
	return err.details
}
func (err *wrapErrorImpl) Error() string {
	return fmt.Sprintf("error code: %v, error msg: %v, error details: %v", err.code, err.msg, err.details)
}
func New(code int, msg string, details ...string) WrapError {
	return &wrapErrorImpl{
		code:    code,
		msg:     msg,
		details: details,
	}
}

func FromErr(err error) WrapError {
	if err == nil {
		return nil
	}
	wrapErr, ok := err.(WrapError)
	if !ok {
		return &wrapErrorImpl{
			code: UnknownError,
			msg:  err.Error(),
		}
	}
	return wrapErr
}
