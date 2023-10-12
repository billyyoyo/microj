package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	"github.com/pkg/errors"
	"testing"
)

type UserSimple struct {
	Name   string `redis:"name"`
	Gender string `redis:"gender"`
	Age    int    `redis:"age"`
}

func (u *UserSimple) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u *UserSimple) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func TestRedis(t *testing.T) {
	config.Init()
	InitRedis()
	cli := Redis()
	ret, err := cli.Incr(context.Background(), "online").Result()
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	fmt.Println("exec ret ", ret)

	user := UserSimple{}
	err = cli.HGetAll(context.Background(), "user1").Scan(&user)
	if err != nil {
		logger.Err(errors.Wrap(err, ""))
		return
	}
	fmt.Printf("%+v\n", user)
}

func TestHSet(t *testing.T) {
	config.Init()
	InitRedis()
	user := UserSimple{
		Name:   "billyyoyo",
		Gender: "female",
		Age:    18,
	}
	b, _ := user.MarshalBinary()
	fmt.Println("------------------", string(b))
	err := Redis().HSet(context.Background(), "users", "master", &user).Err()
	if err != nil {
		panic(err)
	}
}
