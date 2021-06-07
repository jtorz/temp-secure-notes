package serverconfig

import (
	"reflect"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/spf13/viper"
)

type Config struct {
	Port         int              `mapstructure:"PORT"`
	AppMode      string           `mapstructure:"MODE"`
	DataTTL      int              `mapstructure:"DATA_TTL"`
	RedisURL     string           `mapstructure:"REDIS_URL"`
	RedisMaxOpen int              `mapstructure:"REDIS_MAX_OPEN"`
	RedisMaxIdle int              `mapstructure:"REDIS_MAX_IDLE"`
	LoggingLevel config.LogginLvl `mapstructure:"LOGGING_LEVEL"`
}

func LoadConfig() (*Config, error) {
	conf := Config{}
	if config.EnvPrefix != "" {
		viper.SetEnvPrefix(config.EnvPrefix)
	}
	viper.SetTypeByDefaultValue(true)
	RegisterEnvs(conf)

	viper.SetDefault("MODE", "debug")
	viper.SetDefault("DATA_TTL", "3600")
	viper.SetDefault("REDIS_MAX_OPEN", "5")
	viper.SetDefault("REDIS_MAX_IDLE", "5")
	viper.SetDefault("LOGGING_LEVEL", "0")

	err := viper.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func RegisterEnvs(iv interface{}) {
	v := reflect.ValueOf(iv)
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("mapstructure")
		tagValues := strings.Split(tag, ",")
		viper.BindEnv(tagValues[0])
	}
}

func OpenRedis(redisURL string, maxOpen, maxIdle int) (*redis.Pool, error) {
	redis := redis.Pool{
		MaxIdle:     maxOpen,
		MaxActive:   maxIdle,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL, redis.DialTLSSkipVerify(true))
		},
	}

	conn := redis.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}
	return &redis, nil
}
