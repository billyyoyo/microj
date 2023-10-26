package config

import (
	"flag"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/util"
	"github.com/billyyoyo/viper"
	_ "github.com/billyyoyo/viper/remote"
)

var (
	mainYamlPath string
	loader       *viper.Viper
)

func init() {
	flag.StringVar(&mainYamlPath, "c", "app.yml", "main yaml file name")
}

func Init() {
	workspace := util.RunningSpace()
	logger.Info("app run in ", workspace)
	logger.Info("load configs")
	flag.Parse()
	loader = viper.New()
	loader.AutomaticEnv()
	loader.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	loader.AddConfigPath(workspace + "conf")
	// 1. 加载主配置文件 app.yml
	loadLocalConfig(mainYamlPath, false)
	// 2. 加载more-confs的配置文件 如：dev.yml, mysql.yml
	confs := loader.GetStringSlice("config.local.files")
	for _, conf := range confs {
		loadLocalConfig(conf, true)
	}
	// 3. 加载配置中心配置文件  从主配置和active配置中读取 如log.yml
	remoteEnable := loader.GetBool("config.remote.enable")
	if remoteEnable {
		loadRemoteConfig()
	}
}

func loadLocalConfig(confName string, merge bool) {
	loader.SetConfigName(confName)
	var err error
	if merge {
		err = loader.MergeInConfig()
	} else {
		err = loader.ReadInConfig() // Find and read the config file
	}
	if err != nil { // Handle errors reading the config file
		logger.Fatal("load configs failed", err)
	}
	logger.Info("load local config file ", confName)
}

func loadRemoteConfig() {
	keys := loader.GetStringSlice("config.remote.keys")
	prefix := loader.GetString("config.remote.keyPrefix")
	remoteProvider := loader.GetString("config.remote.provider")
	remoteHost := loader.GetString("config.remote.host")
	username := loader.GetString("config.remote.user")
	password := loader.GetString("config.remote.pwd")
	for _, key := range keys {
		loader.AddRemoteProvider(remoteProvider, remoteHost, prefix+key, username, password)
		logger.Info("load remote config file ", key)
	}
	err := loader.ReadRemoteConfig()
	if err != nil {
		logger.Error("load remote config failed", err)
	}
	loader.WatchRemoteConfigOnChannel()
}

func Scan(key string, conf interface{}) error {
	return loader.UnmarshalKey(key, conf)
}

func ScanWithRefresh(en viper.Refreshable) error {
	return loader.UnmarshalWithRefresh(en)
}

func SetDefault(key string, value interface{}) {
	loader.SetDefault(key, value)
}

func GetInt64(key string) int64 {
	return loader.GetInt64(key)
}

func GetInt32(key string) int32 {
	return loader.GetInt32(key)
}

func GetString(key string) string {
	return loader.GetString(key)
}

func GetBool(key string) bool {
	return loader.GetBool(key)
}

func AllConf() map[string]interface{} {
	return loader.AllSettings()
}
