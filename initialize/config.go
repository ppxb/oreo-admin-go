package initialize

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/ppxb/oreo-admin-go/pkg/config"
	"github.com/ppxb/oreo-admin-go/pkg/constant"
	"github.com/ppxb/oreo-admin-go/pkg/global"
	"github.com/ppxb/oreo-admin-go/pkg/log"
	"github.com/ppxb/oreo-admin-go/pkg/utils"
)

const (
	configType            = "yml"
	configDir             = "conf"
	developmentConfig     = "config.dev.yml"
	stagingConfig         = "config.staging.yml"
	productionConfig      = "config.prod.yml"
	defaultConnectTimeout = 5
	defaultEnvPrefix      = "CFG"
	defaultUrlPrefix      = "api"
	defaultApiVersion     = "v1"
)

var (
	sensitiveKeys = []string{
		"CFG_MYSQL_URI",
		"CFG_REDIS_URI",
		"CFG_JWT_REALM",
		"CFG_JWT_KEY",
		"CFG_UPLOAD_OSS_MINIO_SECRET",
	}
)

func Config(ctx context.Context, conf embed.FS) {
	confBox := initConfBox(ctx, conf)
	global.ConfBox = confBox

	v := initViper(confBox)
	loadConfig(confBox, v)

	if err := v.Unmarshal(&global.Conf); err != nil {
		panic(errors.Wrapf(err, "initialize config failed, config env: %s_CONF: %s", global.AppEnvName, confBox.Dir))
	}

	applyEnvOverrides()
	setupLogger()
	normalizeConfig()
	loadRSAKeys(ctx, confBox)

	log.WithContext(ctx).Info("[INIT] Initialize config success, config env: `%s_CONF: /%s`", global.AppEnvName, confBox.Dir)
}

func initConfBox(ctx context.Context, conf embed.FS) config.ConfBox {
	confDir := os.Getenv(fmt.Sprintf("%s_CONF", global.AppEnvName))
	if confDir == "" {
		confDir = configDir
	}

	return config.ConfBox{
		Ctx: ctx,
		Fs:  conf,
		Dir: confDir,
	}
}

func initViper(box config.ConfBox) *viper.Viper {
	v := viper.New()
	readConfig(box, v, developmentConfig)

	for key, val := range v.AllSettings() {
		v.SetDefault(key, val)
	}
	return v
}

func loadConfig(box config.ConfBox, v *viper.Viper) {
	env := strings.ToLower(os.Getenv(fmt.Sprintf("%s_MODE", global.AppProdName)))
	configName := getConfigName(env)
	global.Mode = getMode(env)

	if configName != "" {
		readConfig(box, v, configName)
	}
}

func getConfigName(env string) string {
	switch env {
	case constant.Stage:
		return stagingConfig
	case constant.Prod:
		return productionConfig
	default:
		return ""
	}
}

func getMode(env string) string {
	if env == constant.Stage || env == constant.Prod {
		return env
	}
	return constant.Dev
}

func applyEnvOverrides() {
	envPrefix := strings.ToUpper(os.Getenv(fmt.Sprintf("%s_ENV", global.AppEnvName)))
	if envPrefix == "" {
		envPrefix = defaultEnvPrefix
	}

	utils.EnvToInterface(
		utils.WithEnvObj(&global.Conf),
		utils.WithEnvPrefix(envPrefix),
		utils.WithEnvFormat(formatEnvValue),
	)
}

func formatEnvValue(key string, val interface{}) string {
	if utils.Contains(sensitiveKeys, key) {
		val = "******"
	}
	return fmt.Sprintf("%s: %v", key, val)
}

func setupLogger() {
	log.DefaultWrapper = log.NewWrapper(log.New(
		log.WithCategory(global.Conf.Logs.Category),
		log.WithLevel(global.Conf.Logs.Level),
		log.WithJson(global.Conf.Logs.Json),
		log.WithLineNumPrefix(global.RuntimeRoot),
		log.WithLineNum(!global.Conf.Logs.LineNum.Disable),
		log.WithLineNumLevel(global.Conf.Logs.LineNum.Level),
		log.WithLineNumVersion(global.Conf.Logs.LineNum.Version),
		log.WithLineNumSource(global.Conf.Logs.LineNum.Source),
	))
}

func normalizeConfig() {
	if global.Conf.System.ConnectTimeout < 1 {
		global.Conf.System.ConnectTimeout = defaultConnectTimeout
	}

	if strings.TrimSpace(global.Conf.System.UrlPrefix) == "" {
		global.Conf.System.UrlPrefix = defaultUrlPrefix
	}

	if strings.TrimSpace(global.Conf.System.ApiVersion) == "" {
		global.Conf.System.ApiVersion = defaultApiVersion
	}

	global.Conf.System.Base = fmt.Sprintf("/%s/%s", global.Conf.System.UrlPrefix, global.Conf.System.ApiVersion)

	global.Conf.Mysql.TablePrefix = strings.TrimSuffix(strings.TrimSpace(global.Conf.Mysql.TablePrefix), "_")

	if !global.Conf.Redis.Enable {
		global.Conf.Redis.EnableBinlog = false
	}
}

func loadRSAKeys(ctx context.Context, box config.ConfBox) {
	loadRSAKey(ctx, box, &global.Conf.Jwt.RSAPublicBytes, global.Conf.Jwt.RSAPublicKey, "public")
	loadRSAKey(ctx, box, &global.Conf.Jwt.RSAPrivateBytes, global.Conf.Jwt.RSAPrivateKey, "private")
}

func loadRSAKey(ctx context.Context, box config.ConfBox, target *[]byte, path, keyType string) {
	data := box.Get(path)
	if len(data) == 0 {
		log.WithContext(ctx).Warn("[RSA] Read rsa %s file failed, please check path: %s", keyType, path)
		return
	}
	*target = data
}

func readConfig(box config.ConfBox, v *viper.Viper, configFile string) {
	v.SetConfigType(configType)
	conf := box.Get(configFile)

	if len(conf) == 0 {
		panic(fmt.Sprintf("initialize config failed, config env: `%s_CONF: %s`", global.AppEnvName, box.Dir))
	}

	if err := v.ReadConfig(bytes.NewReader(conf)); err != nil {
		panic(errors.Wrapf(err, "initialize config failed, config env: `%s_CONF: %s`", global.AppEnvName, box.Dir))
	}
}
