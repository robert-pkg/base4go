package errors

import (
	"fmt"
)

var (
	_errMap = map[int32]*Error{} // register codes.
)

func New(e int32, msg string) *Error {
	if e < 200000 {
		panic("must greater than or equal 200000")
	}

	return add(e, msg)
}

func add(e int32, msg string) *Error {
	if _, ok := _errMap[e]; ok {
		panic(fmt.Sprintf("code: %d already exist", e))
	}

	err := &Error{Code: e, Msg: msg}
	_errMap[e] = err
	return err
}

type Error struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (x *Error) GetCode() int32 {
	return x.Code
}

func (x *Error) GetMsg() string {
	return x.Msg
}

func (x *Error) Error() string {
	return x.Msg
}
