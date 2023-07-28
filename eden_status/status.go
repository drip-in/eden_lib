package eden_status

import (
	"errors"
	"fmt"
)

const (
	COMMON  int32 = (1 << 10) * 0 // 0
	SERVICE int32 = (1 << 10) * 1 // 1025
)

// api use
type StatusCode interface {
	Code() int32
	Msg() string
}

type EdenStatus struct {
	StatusCode int32
	StatusMsg  string
	ShowMsgKey string
}

var (
	EdenSuccess = &EdenStatus{StatusCode: COMMON, StatusMsg: ""}
	// stable:stable, code:6001001, starlingKey:, msg:服务器打瞌睡了，请稍后再试。
	EdenServiceInternal = &EdenStatus{StatusCode: COMMON + 1, StatusMsg: "服务器开小差了，请稍后再试"}
	// stable:stable, code:6001002, starlingKey:, msg:参数不合法
	EdenInvalidParam = &EdenStatus{StatusCode: COMMON + 2, StatusMsg: "参数不合法"}
	// stable:stable, code:6001005, starlingKey:, msg:操作频繁，请稍后再试。
	EdenLocked = &EdenStatus{StatusCode: COMMON + 3, StatusMsg: "操作频繁，请稍后再试"}

	//SERVICE
	ERROR_AUTH_CHECK_TOKEN_FAIL    = &EdenStatus{StatusCode: SERVICE + 1, StatusMsg: "认证校验失败"}
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = &EdenStatus{StatusCode: SERVICE + 2, StatusMsg: "认证已过期"}
)

func (this *EdenStatus) GetStatusCode() int32 {
	return this.StatusCode
}

func (this *EdenStatus) GetError() (int32, string) {
	return this.StatusCode, this.StatusMsg
}

// adapt to new eden_status interface
func (this *EdenStatus) Code() int32 {
	return this.StatusCode
}

// adapt to new eden_status
func (this *EdenStatus) Msg() string {
	return this.StatusMsg
}

func NewErrCode(err error) *EdenStatus {
	if err == nil {
		return EdenSuccess
	}
	for {
		if err != nil {
			if errCode, ok := err.(StatusCode); ok {
				return &EdenStatus{
					StatusCode: errCode.Code(),
					StatusMsg:  errCode.Msg(),
				}
			}
			err = errors.Unwrap(err)
		} else {
			break
		}
	}
	return EdenServiceInternal
}

// 替换替换原信息想上透传。
func RepErrCodeMsg(base StatusCode, newMsg string, args ...interface{}) *EdenStatus {
	return &EdenStatus{
		StatusCode: base.Code(),
		StatusMsg:  fmt.Sprintf(newMsg, args...),
	}
}
