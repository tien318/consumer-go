package redis

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var Client *redis.Client
var ClientPersistent *redis.Client

// Init the Client
func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.host"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
}

// InitPersistentRedis init client connect to persistent redis
func InitPersistentRedis() {
	ClientPersistent = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_persistent.host"),
		Password: viper.GetString("redis_persistent.password"),
		DB:       viper.GetInt("redis_persistent.db"),
	})
}
