package query

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm/schema"
)

type RedisOptions struct {
	ctx            context.Context
	redis          redis.UniversalClient
	redisUri       string
	database       string
	namingStrategy schema.Namer
}

func getRedisOptionsOrSetDefault(options *RedisOptions) *RedisOptions {
	if options == nil {
		return &RedisOptions{
			ctx:      context.Background(),
			database: "query_redis",
		}
	}
	return options
}
