package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sync/atomic"
	"testing"
)

func TestRemoveSliceElement(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	tar := 3
	for i := 0; i < len(arr); i++ {
		n := arr[i]
		if n == tar {
			arr = append(arr[:i], arr[i+1:]...)
		}
	}

	fmt.Println(arr)
}

func TestStrBytes(t *testing.T) {
	str := "adf123@#$解锁等级分"
	bs := Str2bytes(str)
	str2 := Bytes2str(bs)
	fmt.Println(str2)
}

func TestSubSlice(t *testing.T) {
	var arr1 []string
	arr2 := []string{"1", "4", "5"}
	fmt.Println(DeleteSlice(arr1, arr2))
	fmt.Println(DeleteSlice(arr2, arr1))
}

func TestQueryParser(t *testing.T) {
	m := map[string]string{
		"name":  "hanjing",
		"age":   "24",
		"count": "123",
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteByte('{')
	var i int32
	for k, v := range m {
		if atomic.LoadInt32(&i) > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.Write([]byte(k))
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.Write([]byte(v))
		buf.WriteByte('"')
		atomic.AddInt32(&i, 1)
	}
	buf.WriteByte('}')
	fmt.Println(buf.String())
	var info struct {
		Name  string `json:"name"`
		Age   int32  `json:"age"`
		Count int64  `json:"count"`
	}
	err := json.Unmarshal(buf.Bytes(), &info)
	//err := json.Unmarshal([]byte("{\"name\":\"hanjing\",\"age\":\"24\",\"count\":\"123\"}"), &info)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(info.Name, info.Age, info.Count)
}

func TestRegexp(t *testing.T) {
	r, _ := regexp.Compile("^/(.)+/(.)+/")
	str := "/service-user/department/tree"
	s := r.FindString(str)
	fmt.Println("'" + s + "'")
}
