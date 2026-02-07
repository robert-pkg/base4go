package errors

import (
	"fmt"
	"testing"
)

func TestEqual(t *testing.T) {
	err := ErrInternalServer
	warpErr := fmt.Errorf("warp %w", err)

	if !Equal(err, warpErr) {
		t.Errorf("err should be equal warpErr")
	}
}

func TestFrom(t *testing.T) {
	err := fmt.Errorf("hello error")

	err2 := FromError(err)

	err3 := fmt.Errorf("warp %w", err2)

	t.Logf("%v\r\n", err)
	t.Logf("code: %v, msg: %v\r\n", err2.GetCode(), err2.GetMsg())
	t.Logf("%v\r\n", err3.Error())

	e := FromError(err3)
	if e != nil {
		t.Logf("code: %v, msg: %v\r\n", e.GetCode(), err3.Error())
	} else {
		t.Log("e is nil")
	}
}
