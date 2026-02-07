package errors

import (
	stderrors "errors"
)

// As finds the first error in err's chain that matches *Error.
func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}

	var merr *Error
	if stderrors.As(err, &merr) {
		return merr, true
	}

	return nil, false
}

// FromError try to convert go error to *Error.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if verr, ok := As(err); ok {
		return verr
	}

	if verr, ok := err.(*Error); ok && verr != nil {
		return verr
	}

	return &Error{
		Code: ErrInternalServer.Code,
		Msg:  err.Error(),
	}
}

func Equal(err1 error, err2 error) bool {
	verr1, ok1 := As(err1)
	verr2, ok2 := As(err2)

	if ok1 != ok2 {
		// 类型不同，肯定不同
		return false
	}

	if !ok1 {
		// 两个都不是 Error
		return stderrors.Is(err1, err2)
	}

	// 两个都是Error，只看Code
	if verr1.Code != verr2.Code {
		return false
	}

	return true
}

func IsSuccess(err error) bool {
	if err == nil {
		return true
	}

	if verr, ok := As(err); ok {
		if verr.Code == E_SUCCESS {
			return true
		}
	}

	return false
}
