package config

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/ppxb/oreo-admin-go/pkg/log"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Env          string `mapstructure:"env"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogLevel        string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

type ConfBox struct {
	Ctx context.Context
	Fs  embed.FS
	Dir string
}

func (c ConfBox) Get(filename string) []byte {
	if filename == "" {
		return nil
	}

	path := c.buildPath(filename)
	if data := c.readFromFileSystem(path); data != nil {
		return data
	}
	return c.readFromEmbed(path)
}

func (c ConfBox) buildPath(filename string) string {
	return filepath.Join(c.Dir, filename)
}

func (c ConfBox) readFromFileSystem(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.WithContext(c.Ctx).WithError(err).Debug("[CONF BOX] Read file %s from file system failed, will try embed", path)
		return nil
	}

	log.WithContext(c.Ctx).Debug("[CONF BOX] Read file %s from file system success", path)
	return data
}

func (c ConfBox) readFromEmbed(path string) []byte {
	data, err := c.Fs.ReadFile(path)
	if err != nil {
		log.WithContext(c.Ctx).WithError(err).Warn("[CONF BOX] Read file %s from embed failed", path)
		return nil
	}

	if len(data) == 0 {
		log.WithContext(c.Ctx).Warn("[CONF BOX] File %s is empty in embed", path)
		return nil
	}

	log.WithContext(c.Ctx).Debug("[CONF BOX] Read file %s from embed success", path)
	return data
}
