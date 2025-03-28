package config

import (
	"sync"

	"github.com/sgostarter/libconfig"
)

type Config struct {
	Listen    string `yaml:"listen"`
	DataRoot  string `yaml:"dataRoot"`
	CacheRoot string `yaml:"cacheRoot"`
}

var (
	_config Config
	_once   sync.Once
)

func GetConfig() *Config {
	_once.Do(func() {
		_, err := libconfig.Load("video-be.yaml", &_config)
		if err != nil {
			panic(err)
		}
	})

	return &_config
}
