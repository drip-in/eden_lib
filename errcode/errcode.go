package errcode

import (
	"errors"
	"fmt"
)

const (
	COMMON  int32 = (1 << 10) * 0 // 0
	SERVICE int32 = (1 << 10) * 1 // 1025
)

var (
	Success = &InnerErrCode{code: COMMON, msg: ""}
	// stable:stable, code:6001001, starlingKey:, msg:服务器打瞌睡了，请稍后再试。
	ErrServiceInternal = &InnerErrCode{code: COMMON + 1, msg: "服务器打瞌睡了，请稍后再试"}
	// stable:stable, code:6001002, starlingKey:, msg:参数不合法
	ErrInvalidParam = &InnerErrCode{code: COMMON + 2, msg: "参数不合法"}
	// stable:stable, code:6001005, starlingKey:, msg:操作频繁，请稍后再试。
	ErrLocked = &InnerErrCode{code: COMMON + 3, msg: "操作频繁，请稍后再试"}

	//SERVICE
	ERR_AUTH_CHECK_TOKEN_FAIL    = &InnerErrCode{code: SERVICE + 1, msg: "认证校验失败"}
	ERR_AUTH_CHECK_TOKEN_TIMEOUT = &InnerErrCode{code: SERVICE + 2, msg: "认证已过期"}
)

// rpc use
type ErrCode interface {
	Code() int32
	Error() string
}

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

func (i *InnerErrCode) Error() string {
	return i.msg
}

//func NewErrCode(err error) ErrCode {
//	if err == nil {
//		return Success
//	}
//	for {
//		if err != nil {
//			if errCode, ok := err.(ErrCode); ok {
//				return errCode
//			}
//			err = errors.Unwrap(err)
//		} else {
//			break
//		}
//	}
//	return ErrServiceInternal
//}

// 扩展错误msg向上传递。如"参数不合法： xxx"。 xxx即为扩展信息
func ExtErrCodeMsg(base ErrCode, extMsg string, args ...interface{}) error {
	return &InnerErrCode{
		code: base.Code(),
		msg:  fmt.Sprintf("%v: %v", base.Error(), fmt.Sprintf(extMsg, args...)),
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

