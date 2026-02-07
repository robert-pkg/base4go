package net

import (
	"fmt"
	"math/rand"
	"syscall"
	"time"

	"github.com/robert-pkg/base4go/log"
)

// RandPort 生成随机端口号 (10240, 65535), 保证可用
func RandPort(min, max int, maxTryCnt int) (int, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for tryCnt := 0; tryCnt < maxTryCnt; tryCnt++ {
		sa := new(syscall.SockaddrInet4)
		sa.Port = r.Intn(max-min+1) + min
		sa.Addr = [4]byte{0x7f, 0x00, 0x00, 0x01} // 其实就是127.0.0.1
		if s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); nil != err {
			log.Warnf("warn: %v", err)
		} else {
			if err := syscall.Bind(s, sa); err != nil {
				log.Infof("bind fail, port: %d, err: %v", sa.Port, err)
			} else {
				syscall.Close(s)
				return sa.Port, nil
			}
		}

	}

	return 0, fmt.Errorf("RandPort fail")
}
