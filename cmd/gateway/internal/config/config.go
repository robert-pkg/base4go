package config

import (
	"errors"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/robert-pkg/base4go/cmd/gateway/internal/utils"
	"github.com/robert-pkg/base4go/log"
	"github.com/spf13/viper"
)

const (
	default_env_prefix = "gateway"
)

var (
	globalConfig *Config
)

type Server struct {
	Port      int    `mapstructure:"port"`
	ApiPrefix string `mapstructure:"api_prefix"`
}

type Config struct {
	Server Server

	Registry struct {
		Addr string `mapstructure:"addr"`
	}

	WarmServer struct {
		Grpc []string `mapstructure:"grpc"` // 预热的 grpc 服务
	}

	apiMapping map[string]*apiMappintItem
	ApiMapping map[string]string `mapstructure:"api_mapping"`
}

func (c *Config) check() error {
	if c.Server.Port == 0 {
		return errors.New("server.port is required")
	}

	if len(c.Server.ApiPrefix) == 0 {
		c.Server.ApiPrefix = "/api/base4go"
	}

	if len(c.Registry.Addr) == 0 {
		c.Registry.Addr = "127.0.0.1:8500"
	}

	if len(c.ApiMapping) == 0 {
		c.ApiMapping = map[string]string{}
	}

	c.apiMapping = make(map[string]*apiMappintItem)
	for k, v := range c.ApiMapping {
		scheme, service, method, err := utils.ParseAddr(v)
		if err != nil {
			log.Error("err", "err", err)
			return err
		}

		c.apiMapping[k] = &apiMappintItem{
			Scheme:  scheme,
			Service: service,
			Method:  method,
		}
	}

	return nil
}

// Viper 的优先级：Set > Env > Config File > Default
func LoadConfig(yamlPath string, envPrefix string) (*Config, error) {
	viper.SetConfigFile(yamlPath)

	usedEnvPrefix := default_env_prefix
	if len(envPrefix) > 0 {
		usedEnvPrefix = envPrefix
	}

	// 允许环境变量覆盖
	viper.SetEnvPrefix(usedEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 默认值
	viper.SetDefault("server.port", 8080)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if err = cfg.check(); err != nil {
		return nil, err
	}

	globalConfig = cfg

	setApiMapping(globalConfig.apiMapping)

	// 👇 开启热加载
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		newCfg := &Config{}
		if err := viper.Unmarshal(newCfg); err != nil {
			log.Errorf("config reload failed: %v", err)
			return
		}

		if err := newCfg.check(); err != nil {
			log.Errorf("config reload failed: %v", err)
			return
		}

		hotReload(newCfg)
	})

	return globalConfig, nil
}
