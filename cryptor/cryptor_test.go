package cryptor

import (
	"fmt"
	"github.com/billyyoyo/microj/logger"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm4"
	"testing"
)

func TestSM3(t *testing.T) {
	fmt.Println(Sum("hanjing", "salt"))
}

func TestSM4(t *testing.T) {
	key := "timestack2021131"
	iv := "0000000000000000"
	text := "hello world特点：将前一个密文块作为输入来生成密钥流，然后与明文块进行异或操作，可以实现流密码的功能。"
	cypher, err := Encrypt(key, text, iv)
	if err != nil {
		logger.Err(err)
		return
	}
	fmt.Println(cypher)
	plain, err := Decrypt(key, cypher, iv)
	if err != nil {
		logger.Err(err)
		return
	}
	fmt.Println(plain)
}

func TestSM(t *testing.T) {
	plain := "hello world"
	key := "timestack2021131"
	//iv := "0000000000000000"
	//if iv != "" {
	//	err := sm4.SetIV([]byte(iv))
	//	if err != nil {
	//		err = errors.Wrap(err, "")
	//		return
	//	}
	//}
	ecbMsg, err := sm4.Sm4Ecb([]byte(key), []byte(plain), true)
	if err != nil {
		err = errors.Wrap(err, "")
		return
	}
	fmt.Printf("c: %x\n", ecbMsg)
	fmt.Printf("s: %s\n", string(ecbMsg))
	ecbDec, err := sm4.Sm4Ecb([]byte(key), ecbMsg, false) //sm4Ecb模式pksc7填充解密
	if err != nil {
		err = errors.Wrap(err, "")
		return
	}
	fmt.Printf("d: %x\n", ecbDec)

}
