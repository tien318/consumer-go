package redis

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// Client is Redis Client instance
var Client *redis.Client

// Init the Client
func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.host"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
}
