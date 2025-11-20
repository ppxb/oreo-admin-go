package main

import (
	"embed"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"

	"github.com/ppxb/oreo-admin-go/initialize"
	"github.com/ppxb/oreo-admin-go/pkg/global"
	"github.com/ppxb/oreo-admin-go/pkg/log"
	"github.com/ppxb/oreo-admin-go/pkg/tracing"
)

//go:embed conf
var conf embed.FS

func main() {
	ctx := tracing.NewId(nil)

	defer func() {
		if err := recover(); err != nil {
			log.WithContext(ctx).WithError(errors.Errorf("%v", err)).Error("[SERVER] Failed to start server, stack is: %s", string(debug.Stack()))
		}
	}()

	_, file, _, _ := runtime.Caller(0)
	global.RuntimeRoot = strings.TrimSuffix(file, "main.go")

	initialize.Config(ctx, conf)
	initialize.Mysql(ctx)
}
