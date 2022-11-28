package el_utils

import (
	"github.com/drip-in/eden_lib/logs"
	"time"
)

func Retry(count int, sleep int, f func() (success bool)) bool {
	for retry := 0; retry < count; retry++ {
		success := f()
		if success {
			return true
		} else {
			left := count - retry - 1
			if left == 0 {
				return false
			} else {
				logs.Warn("[Retry]", logs.Int("sleep", sleep), logs.Int("left", left))
				time.Sleep(time.Duration(sleep) * time.Second)
			}
		}
	}
	return false
}
