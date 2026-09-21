package ioc

import (
	"webook/config"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: config.AppConfig.Redis.Addr,
	})
}

func InitRedisCmdable(client *redis.Client) redis.Cmdable {
	return client
}
