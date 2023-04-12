package util

import (
	"fmt"
	"github.com/billyyoyo/microj/errs"
	"github.com/gin-gonic/gin/binding"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"unsafe"
)

func RunningSpace() string {
	runPath, _ := exec.LookPath(os.Args[0])
	if strings.HasSuffix(runPath, ".test") {
		runPath = fmt.Sprintf("$PWD%s..%s", string(os.PathSeparator), string(os.PathSeparator))
	} else {
		runPath = ""
	}
	return runPath
}

func GetIP() string {
	ifaces, _ := net.Interfaces()

	for i := 0; i < len(ifaces); i++ {
		addrs, _ := ifaces[i].Addrs()

		for j := 0; j < len(addrs); j++ {
			addr := addrs[j]
			if ip, ok := addr.(*net.IPNet); ok && ip.IP.IsLoopback() {
				continue
			}

			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP.String()
			case *net.IPAddr:
				return v.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func IF(cond bool, val1, val2 interface{}) interface{} {
	if cond {
		return val1
	}
	return val2
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func DeleteSlice(arr []string, sub []string) (left []string) {
	m := map[string]bool{}
	for _, s := range sub {
		m[s] = true
	}
	for _, a := range arr {
		if _, ok := m[a]; !ok {
			left = append(left, a)
		}
	}
	return
}

func QueryUnmarshal(data []byte, ptr any) error {
	v, err := url.ParseQuery(Bytes2str(data))
	if err != nil {
		return errs.Wrap(errs.ERRCODE_GATEWAY, "parse query params error", err)
	}
	err = binding.MapFormWithTag(ptr, v, "form")
	if err != nil {
		return errs.Wrap(errs.ERRCODE_GATEWAY, "bind query params error", err)
	}
	return nil
}
