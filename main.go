package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/ppxb/oreo-admin-go/internal/router"
	"github.com/ppxb/oreo-admin-go/pkg/config"
	"github.com/ppxb/oreo-admin-go/pkg/global"
	"github.com/ppxb/oreo-admin-go/pkg/log"
	"github.com/ppxb/oreo-admin-go/pkg/tracing"
)

func main() {
	var ctx = tracing.NewId(nil)

	defer func() {
		if err := recover(); err != nil {
			log.WithContext(ctx).WithError(errors.Errorf("%v", err)).Error("[SERVER] Failed to start server, stack is: %s", string(debug.Stack()))
		}
	}()

	_, file, _, _ := runtime.Caller(0)
	global.RuntimeRoot = strings.TrimSuffix(file, "main.go")

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("[CONFIG] Failed to load configuration")
	}

	logger := log.WithContext(ctx)

	// db, err := database.InitMySQL(cfg.MySQL)
	// if err != nil {
	// 	log.Fatal("初始化数据库失败", zap.Error(err))
	// }
	// sqlDB, _ := db.DB()
	// defer sqlDB.Close()

	r := router.NewRouter(cfg)

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Info("[INIT] Starting Oreo Admin Server...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("[SERVER] Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("[SHUTDOWN] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("[SHUTDOWN] Server forced to shutdown")
	}

	logger.Info("[SHUTDOWN] Server exited")
}
