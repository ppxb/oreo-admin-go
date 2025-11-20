package initialize

import (
	"context"
	"embed"
	"fmt"
	"time"

	m "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	glogger "gorm.io/gorm/logger"

	"github.com/ppxb/oreo-admin-go/pkg/global"
	"github.com/ppxb/oreo-admin-go/pkg/log"
	"github.com/ppxb/oreo-admin-go/pkg/migrate"
)

//go:embed db/*.sql
var sqlFs embed.FS

func Mysql(ctx context.Context) {
	if err := parseDSNConfig(); err != nil {
		panic(errors.Wrap(err, "initialize mysql failed"))
	}

	if err := executeMigration(ctx); err != nil {
		panic(errors.Wrap(err, "mysql migration failed"))
	}

	log.WithContext(ctx).Info("[INIT] Mysql initialized successfully")
}

func parseDSNConfig() error {
	cfg, err := m.ParseDSN(global.Conf.Mysql.Uri)
	if err != nil {
		return err
	}
	global.Conf.Mysql.DSN = *cfg
	return nil
}

func executeMigration(ctx context.Context) error {
	return migrate.Do(
		migrate.WithCtx(ctx),
		migrate.WithUri(global.Conf.Mysql.Uri),
		migrate.WithFs(sqlFs),
		migrate.WithFsRoot("db"),
		migrate.WithBefore(initializeDatabase),
	)
}

func initializeDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Conf.System.ConnectTimeout)*time.Second)
	defer cancel()

	go monitorTimeout(ctx)

	db, err := createDBConnection()
	if err != nil {
		return err
	}

	global.Mysql = db
	return nil
}

func createDBConnection() (*gorm.DB, error) {
	l := log.NewDefaultGormLogger()
	if global.Conf.Mysql.NoSql {
		l = l.LogMode(glogger.Silent)
	} else {
		l = l.LogMode(glogger.Info)
	}

	return gorm.Open(mysql.Open(global.Conf.Mysql.DSN.FormatDSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   global.Conf.Mysql.TablePrefix + "_",
			SingularTable: true,
		},
		QueryFields: true,
		Logger:      l,
	})
}

func monitorTimeout(ctx context.Context) {
	<-ctx.Done()
	if global.Mysql == nil {
		panic(fmt.Sprintf("mysql connection timeout after %d seconds", global.Conf.System.ConnectTimeout))
	}
}
