package redis

import (
	config "nas_server/conf"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func RedisInit(redisConfig *config.RedisConfig) {
	client = redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
		Password: redisConfig.Password,
		DB: redisConfig.DB,
	})
}

func GetClient() *redis.Client {
	return client
}

func CheckNil(err error) bool {
	return err == redis.Nil
}