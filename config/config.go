package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	DB      DBConfig      `mapstructure:"db"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Swagger SwaggerConfig `mapstructure:"swagger"`
}

// AppConfig 应用配置
type AppConfig struct {
	Addr string `mapstructure:"addr"`
	Mode string `mapstructure:"mode"`
}

// DBConfig 数据库配置
type DBConfig struct {
	DSN string `mapstructure:"dsn"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	Salt               string `mapstructure:"salt"`
	ExpireDays         int    `mapstructure:"expire_days"`
	RenewThresholdDays int    `mapstructure:"renew_threshold_days"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SwaggerConfig Swagger 配置
type SwaggerConfig struct {
	Enable bool `mapstructure:"enable"`
}

// Load 读取配置文件
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvPrefix("MJ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}
	return &cfg, nil
}

// MustLoad 读取配置文件（失败直接 panic）
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.addr", ":8080")
	v.SetDefault("app.mode", "debug")
	v.SetDefault("swagger.enable", true)
	v.SetDefault("jwt.expire_days", 3)
	v.SetDefault("jwt.renew_threshold_days", 2)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
}
