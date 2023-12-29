package redis

import (
	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func RedisInit(addr, password string, db int) {
	client = redis.NewClient(&redis.Options{
		Addr: addr,
		Password: password,
		DB: db,
	})
}

func GetClient() *redis.Client {
	return client
}

func CheckNil(err error) bool {
	return err == redis.Nil
}