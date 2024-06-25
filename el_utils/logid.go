package el_utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/drip-in/eden_lib/internal"
	"strings"
	"time"
)

var (
	localIP           string
	fullLengthLocalIP []byte
)

func init() {
	localIP = internal.LocalIP()
	elements := strings.Split(localIP, ".")
	for i := 0; i < len(elements); i++ {
		elements[i] = fmt.Sprintf("%03s", elements[i])
	}
	fullLengthLocalIP = []byte(strings.Join(elements, ""))
}

// genLogId generates a global unique log id for request
// format: %Y%m%d%H%M%S + ip + 5位随机数
// python runtime使用的random uuid, 这里简单使用random产生一个5位数字随机串
func GenLogId() string {
	buf := make([]byte, 0, 64)
	buf = time.Now().AppendFormat(buf, "20060102150405")
	buf = append(buf, fullLengthLocalIP...)

	uuidBuf := make([]byte, 4)
	_, err := rand.Read(uuidBuf)
	if err != nil {
		panic(err)
	}
	uuidNum := binary.BigEndian.Uint32(uuidBuf)
	buf = append(buf, fmt.Sprintf("%05d", uuidNum)[:5]...)
	return string(buf)
}
