package registry

import (
	"fmt"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	opts := Options{
		Host:        "localhost:2379",
		ServiceName: "server-test",
		Weight:      1,
		Enable:      true,
		Port:        8080,
		Timeout:     5,
	}
	Init(opts)
	timer1 := time.Tick(5 * time.Second)
	ch := make(chan bool)
	for {
		select {
		case <-timer1:
			fmt.Println("---------------------------------------------------")
			names, err := ListServices()
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			// todo need avoid 2 inner range loop code style
			for _, name := range names {
				fmt.Println(name)
				service, errr := GetService(name)
				if errr != nil {
					fmt.Println(errr.Error())
					continue
				}
				for _, node := range service.Nodes {
					fmt.Println("\t", node)
				}
			}
		case <-ch:
			return
		}
	}
}
