package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"github.com/billyyoyo/microj/registry"
	"github.com/billyyoyo/microj/util"
	"github.com/go-resty/resty/v2"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRest(t *testing.T) {
	cli := resty.New()
	cli.SetBaseURL("lb://server-api")
	cli.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		if strings.HasPrefix(c.BaseURL, "lb://") {
			req.URL = fmt.Sprintf("http://localhost:8003%s", req.URL)
		}
		return nil
	})
	resp, err := cli.R().
		SetHeader("Authorization", "abc").
		SetQueryParam("id", "10110").
		Get("/info")
	if err != nil {
		logger.Error("call failed", err)
		return
	}
	fmt.Println(resp.String())
}

func TestRestLB(t *testing.T) {
	tick := time.Tick(time.Second)
	var i int64
	for t := range tick {
		i++
		cli := resty.New()
		resp, err := cli.R().SetHeader("Authorization", "1010101010").
			SetQueryParam("id", "10101").
			SetQueryParam("nonce", strconv.FormatInt(t.UnixMilli(), 10)).
			Get("http://localhost:8005/user")
		if err != nil {
			fmt.Println(i, "error", err.Error())
		} else {
			fmt.Println(i, resp.StatusCode(), resp.Status(), util.Bytes2str(resp.Body()))
		}
	}
}

func TestRpcLB(t *testing.T) {
	opts := registry.Options{
		Host:        "localhost:2379",
		ServiceName: "server-test",
		Weight:      1,
		Enable:      true,
		Port:        8080,
		Timeout:     5,
	}
	registry.Init(opts)
	timer1 := time.Tick(time.Second)
	conn := NewRpcConn("server-proto")
	if conn == nil {
		return
	}
	cli := proto.NewExampleClient(conn)
	for {
		select {
		case t := <-timer1:
			req := &proto.Request{
				Value: "hello world",
			}
			resp, err := cli.Call(context.Background(), req)
			if err != nil {
				fmt.Println(t.Format(time.DateTime), err.Error())
				break
			} else {
				fmt.Println(t.Format(time.DateTime), resp.GetMsg())
			}
		}
	}
}

func TestFast(t *testing.T) {
	addrs := []string{
		"localhost:8003",
		"localhost:8004",
	}
	var lbc fasthttp.LBClient
	for _, addr := range addrs {
		c := &fasthttp.HostClient{
			Addr: addr,
		}
		lbc.Clients = append(lbc.Clients, c)
	}
	var req fasthttp.Request
	var resp fasthttp.Response
	for i := 0; i < 10; i++ {
		url := "api://localhost:8080/login"
		req.SetRequestURI(url)
		req.Header.SetMethod(fasthttp.MethodPost)
		bs, _ := json.Marshal(map[string]string{
			"username": "billyyoyo",
			"password": "123123123",
		})
		req.AppendBody(bs)
		if err := lbc.Do(&req, &resp); err != nil {
			logger.Error("error when sending request", err)
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			logger.Error("unexpected status code: ", errors.New(fmt.Sprintf("http error %d", resp.StatusCode())))
		}
	}
}
