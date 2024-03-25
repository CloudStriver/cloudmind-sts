package config

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type EmailConf struct {
	Host     string
	Port     int32
	Password string
	Email    string
}

type CosConfig struct {
	AppId      string
	BucketName string
	Region     string
	SecretId   string
	SecretKey  string
}

type CDNConfig struct {
	Url    string
	Key    string
	Prefix string
	MinTTL int
	MaxTTL int
}

type FilterConfig struct {
	UseDuplicateWordFilter        bool
	UseDBCcaseConverter           bool
	UseSimplifiedChineseConverter bool
	UseIgnoreCase                 bool
}

func (c *CosConfig) CosHost() string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", c.BucketName, c.Region)
}

type Config struct {
	service.ServiceConf
	ListenOn string
	Mongo    struct {
		URL string
		DB  string
	}
	CacheConf     cache.CacheConf
	Redis         *redis.RedisConf
	EmailConf     EmailConf
	CosConfig     *CosConfig
	FileCosConfig *CosConfig
	CdnConfig     *CDNConfig
	FilterConfig  *FilterConfig
}

func NewConfig() (*Config, error) {
	c := new(Config)
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/config.yaml"
	}
	err := conf.Load(path, c)
	if err != nil {
		return nil, err
	}
	err = c.SetUp()
	if err != nil {
		return nil, err
	}
	return c, nil
}
