package app

import (
	"github.com/robert-pkg/base4go/errors"
)

type BaseResp[T any] interface {
	SetCode(int32)
	SetMsg(string)
	SetData(T)
}

func Success[T any, R BaseResp[T]](resp R, data T) R {
	resp.SetCode(0)
	resp.SetMsg("success")
	resp.SetData(data)
	return resp
}

func Fail[T any, R BaseResp[T]](resp R, code int32, msg string) R {
	resp.SetCode(code)
	resp.SetMsg(msg)
	return resp
}

func MakeResp[T any, R BaseResp[T]](err error, resp R, data T) R {
	e := errors.FromError(err)
	if e == nil {
		return Success(resp, data)
	}

	resp.SetCode(e.GetCode())
	resp.SetMsg(err.Error())
	return resp
}
