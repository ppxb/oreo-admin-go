package query

import (
	"context"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ppxb/oreo-admin-go/pkg/tracing"
)

type Redis struct {
	ops        RedisOptions
	Ctx        context.Context
	Error      error
	clone      int
	Statement  *gorm.Statement
	cacheStore *sync.Map
}

func NewRedis(options ...func(*RedisOptions)) Redis {
	ops := getRedisOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.redis == nil {
		if ops.redisUri != "" {
			var err error
			ops.redis, err = ParseRedisURI(ops.redisUri)
			if err != nil {
				panic(err)
			}
		} else {
			panic("redis client is empty")
		}
	}
	if ops.namingStrategy == nil {
		panic("redis namingStrategy is empty")
	}
	rds := Redis{
		ops:   *ops,
		clone: 1,
	}
	rdsCtx := tracing.NewId(ops.ctx)
	rds.Ctx = rdsCtx
	return rds
}

func ParseRedisURI(uri string) (client redis.UniversalClient, err error) {
	var opt asynq.RedisConnOpt
	if uri != "" {
		opt, err = asynq.ParseRedisURI(uri)
		if err != nil {
			return
		}
		client = opt.MakeRedisClient().(redis.UniversalClient)
		return
	}
	err = errors.Errorf("invalid redis config")
	return
}
