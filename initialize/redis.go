package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ppxb/oreo-admin-go/pkg/global"
	"github.com/ppxb/oreo-admin-go/pkg/log"
	"github.com/ppxb/oreo-admin-go/pkg/query"
)

// TODO: NEED REFACTOR(NO CONNECTION FAILED DESCRIPTION LIKE PASSWORD ERROR OR ETC.)

func Redis(ctx context.Context) {
	if !global.Conf.Redis.Enable {
		log.WithContext(ctx).Info("[INIT] Redis is not enabled")
		return
	}
	init := false
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Conf.System.ConnectTimeout)*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				if !init {
					panic(fmt.Sprintf("initialize redis failed: connect timeout(%ds)", global.Conf.System.ConnectTimeout))
				}
				return
			}
		}
	}()

	client, err := query.ParseRedisURI(global.Conf.Redis.Uri)
	if err != nil {
		panic(errors.Wrap(err, "initialize redis failed"))
	}
	err = client.Ping(ctx).Err()
	if err != nil {
		panic(errors.Wrap(err, "initialize redis failed"))
	}
	global.Redis = client

	init = true
	log.WithContext(ctx).Info("[INIT] Initialize redis successfully")
}
