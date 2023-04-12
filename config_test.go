package microj

import (
	"flag"
	"fmt"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	"github.com/billyyoyo/viper"
	"testing"
)

type MysqlConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
	Db   string `yaml:"db"`
	User string `yaml:"user"`
	Pwd  string `yaml:"pwd"`
}

func (w MysqlConfig) CanRefresh() bool {
	return true
}

func (w MysqlConfig) KeyName() string {
	return "mysql"
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
	Auth string `yaml:"auth"`
}

func (w RedisConfig) CanRefresh() bool {
	return true
}

func (w RedisConfig) KeyName() string {
	return "redis"
}

type UploadConfig struct {
	Path    string `yaml:"path"`
	MaxSize int64  `yaml:"maxSize"`
	Timeout int64  `yaml:"timeout"`
}

func (w UploadConfig) CanRefresh() bool {
	return true
}

func (w UploadConfig) KeyName() string {
	return "upload"
}

type GatewayConfig struct {
	Timeout   int64    `yaml:"timeout"`
	Routes    []Route  `yaml:"route" mapstructure:"route"`
	WhiteList []string `yaml:"white-list" mapstructure:"white-list"`
}
type Route struct {
	Id     string `yaml:"id"`
	Path   string `yaml:"path"`
	Schema string `yaml:"schema"`
}

func (g GatewayConfig) CanRefresh() bool {
	return true
}

func (g GatewayConfig) KeyName() string {
	return "gateway"
}

func TestLoadLocal(t *testing.T) {
	viper.SetConfigFile("yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/home/billyyoyo/workspace/microj/conf/")
	viper.SetConfigName("gateway.yml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c := GatewayConfig{}
	err = viper.UnmarshalKey("gateway", &c)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(c.Timeout)
	fmt.Println(c.Routes)
	fmt.Println(c.WhiteList)
}

func TestLoadConfig(t *testing.T) {
	var err error
	flag.Parse()
	config.Init()
	mysqlConf := MysqlConfig{}
	redisConf := RedisConfig{}
	uploadConf := UploadConfig{}
	gatewayConf := GatewayConfig{}

	if err = config.ScanWithRefresh(&mysqlConf); err != nil {
		logger.Error("load mysql config failed", err)
	}
	if err = config.ScanWithRefresh(&redisConf); err != nil {
		logger.Error("load redis config failed", err)
	}
	if err = config.ScanWithRefresh(&uploadConf); err != nil {
		logger.Error("load upload config failed", err)
	}
	if err = config.ScanWithRefresh(&gatewayConf); err != nil {
		logger.Error("load upload config failed", err)
	}

	fmt.Println(gatewayConf.WhiteList)

	//go func(cs ...viper.Refreshable) {
	//	for {
	//		for _, c := range cs {
	//			json, _ := json2.Marshal(c)
	//			logger.Info(c.KeyName(), util.Bytes2str(json))
	//		}
	//		time.Sleep(100 * time.Millisecond)
	//	}
	//
	//}(&mysqlConf, &redisConf, &uploadConf)
	//
	//running := make(chan int)
	//<-running
}
