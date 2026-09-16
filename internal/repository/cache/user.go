package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"webbook/internal/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = redis.Nil

type UserCache struct {
	// 可以传单体 redis.Client，也可以传集群 redis.ClusterClient
	client     redis.Cmdable
	expiration time.Duration
}

// A 用到了 B, B 一定是接口
// A 用到了 B, B 一定是 A 的字段
// A 用到了 B, A 绝对不初始化 B， 而是外部注入
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 5,
	}
}

func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal(val, &user)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
