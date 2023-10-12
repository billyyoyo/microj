package db

import (
	"context"
	"fmt"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

var (
	masterdb  *redis.Client
	clusterdb *redis.ClusterClient
	rConfig   *RConfig
)

type RConfig struct {
	Mode       string      `yaml:"mode"`
	Debug      bool        `yaml:"debug"`
	Address    []string    `yaml:"addr" mapstructure:"addr"`
	Password   string      `yaml:"password"`
	MasterName string      `yaml:"masterName"`
	Pool       RPoolConfig `yaml:"pool" mapstructure:"pool"`
}

type RPoolConfig struct {
	Size        int   `yaml:"size"`
	MinIdle     int   `yaml:"minIdleSize"`
	MaxAge      int64 `yaml:"maxAge"`
	PoolTimeout int64 `yaml:"poolTimeout"`
	IdleTimeout int64 `yaml:"idleTimeout"`
}

func InitRedis() {
	err := config.Scan("redis", &rConfig)
	if err != nil {
		logger.Fatal("redis config load failed", errors.Wrap(err, ""))
	}
	if len(rConfig.Address) == 0 {
		logger.Fatal("no redis address config", errors.New("no redis address config"))
	}

	redis.SetLogger(&rLogger{})

	if rConfig.Mode == "sentinel" {
		if rConfig.MasterName == "" {
			logger.Fatal("no redis master name config", errors.New("no redis master name config"))
		}
		opts := &redis.FailoverOptions{
			MasterName:    rConfig.MasterName,
			SentinelAddrs: rConfig.Address,
			Password:      rConfig.Password,
			SlaveOnly:     false, // 只访问从节点，读写分离
			RouteRandomly: true,  // 只读指令路由到主和从节点
		}
		//if rConfig.Debug {
		//	opts.OnConnect = onConnect
		//}
		if rConfig.Pool.Size > 0 {
			opts.PoolSize = rConfig.Pool.Size
		}
		if rConfig.Pool.MinIdle > 0 {
			opts.MinIdleConns = rConfig.Pool.MinIdle
		}
		if rConfig.Pool.MaxAge > 0 {
			opts.MaxConnAge = time.Duration(rConfig.Pool.MaxAge) * time.Second
		}
		if rConfig.Pool.IdleTimeout > 0 {
			opts.IdleTimeout = time.Duration(rConfig.Pool.IdleTimeout) * time.Second
		}
		if rConfig.Pool.PoolTimeout > 0 {
			opts.PoolTimeout = time.Duration(rConfig.Pool.PoolTimeout) * time.Second
		}
		clusterdb = redis.NewFailoverClusterClient(opts)
	} else {
		opts := &redis.Options{
			Addr:     rConfig.Address[0],
			Password: rConfig.Password,
		}
		//if rConfig.Debug {
		//	opts.OnConnect = onConnect
		//}
		if rConfig.Pool.Size > 0 {
			opts.PoolSize = rConfig.Pool.Size
		}
		if rConfig.Pool.MinIdle > 0 {
			opts.MinIdleConns = rConfig.Pool.MinIdle
		}
		if rConfig.Pool.MaxAge > 0 {
			opts.MaxConnAge = time.Duration(rConfig.Pool.MaxAge) * time.Second
		}
		if rConfig.Pool.IdleTimeout > 0 {
			opts.IdleTimeout = time.Duration(rConfig.Pool.IdleTimeout) * time.Second
		}
		if rConfig.Pool.PoolTimeout > 0 {
			opts.PoolTimeout = time.Duration(rConfig.Pool.PoolTimeout) * time.Second
		}
		masterdb = redis.NewClient(opts)
	}

}

func Redis() redis.Cmdable {
	if rConfig.Mode == "sentinel" {
		return clusterdb
	} else {
		return masterdb
	}
}

func onConnect(ctx context.Context, conn *redis.Conn) error {
	fmt.Printf("%+v\n", conn)
	return nil
}

type rLogger struct{}

func (l *rLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	logger.Infof(format, v...)
}
