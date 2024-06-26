package errcode

import (
	"errors"
	"fmt"
)

// api use
type StatusCode interface {
	Code() int32
	Msg(language string) string
	StarlingKey() string
}

type Stable interface {
	Stability() int32
}

// rpc use
type ErrCode interface {
	Code() int32
	Msg() string
}

const (
	ErrCodeStable   = 0
	ErrCodeUnStable = 1
)

const (
	COMMON int32 = 0
)

var (
	Success            = &InnerErrCode{code: COMMON, msg: ""}
	ErrServiceInternal = &InnerErrCode{code: COMMON + 1, msg: "服务器开小差了，请稍后再试"}
	ErrInvalidParam    = &InnerErrCode{code: COMMON + 2, msg: "参数不合法"}
	ErrLocked          = &InnerErrCode{code: COMMON + 3, msg: "操作频繁，请稍后再试"}
)

type InnerErrCode struct {
	code int32
	msg  string
}

func NewInnerErrCode(code int32, msg string) *InnerErrCode {
	return &InnerErrCode{code: code, msg: msg}
}

func (i *InnerErrCode) Code() int32 {
	return i.code
}

func (i *InnerErrCode) Msg() string {
	return i.msg
}

func (i *InnerErrCode) Error() string {
	return i.msg
}

// 扩展错误msg向上传递。如"参数不合法： xxx"。 xxx即为扩展信息
func ExtErrCodeMsg(base ErrCode, extMsg string, args ...interface{}) error {
	return &InnerErrCode{
		code: base.Code(),
		msg:  fmt.Sprintf("%v: %v", base.Msg(), fmt.Sprintf(extMsg, args...)),
	}
}

// 替换替换原信息想上透传。
func RepErrCodeMsg(base ErrCode, newMsg string, args ...interface{}) error {
	return &InnerErrCode{
		code: base.Code(),
		msg:  fmt.Sprintf(newMsg, args...),
	}
}

// server向外抛错误时用到
func IsLegacyErr(err error) bool {
	if err == nil {
		return false
	}
	for {
		if err != nil {
			if _, ok := err.(ErrCode); ok {
				return true
			}
			err = errors.Unwrap(err)
		} else {
			break
		}
	}
	return false
}
