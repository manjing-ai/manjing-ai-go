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
	Storage StorageConfig `mapstructure:"storage"`
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

// StorageConfig 存储配置
type StorageConfig struct {
	Type           string       `mapstructure:"type"` // local/cos
	MaxFileSizeMB  int64        `mapstructure:"max_file_size_mb"`
	MaxTotalSizeMB int64        `mapstructure:"max_total_size_mb"`
	Local          LocalStorage `mapstructure:"local"`
	COS            COSStorage   `mapstructure:"cos"`
}

// LocalStorage 本地存储
type LocalStorage struct {
	BaseDir string `mapstructure:"base_dir"`
	BaseURL string `mapstructure:"base_url"`
}

// COSStorage 腾讯云 COS
type COSStorage struct {
	Bucket string `mapstructure:"bucket"`
	Region string `mapstructure:"region"`
}

// SwaggerConfig Swagger 配置
type SwaggerConfig struct {
	Enable bool `mapstructure:"enable"`
}

// Load 读取配置文件
func Load() (*Config, error) {
	return LoadWithPath("")
}

// LoadWithPath 指定配置文件路径读取
func LoadWithPath(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvPrefix("MJ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	env := v.GetString("env")
	if env == "" {
		env = "dev"
	}
	v.Set("env", env)

	// 指定路径优先
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config failed: %w", err)
		}
	} else {
		// 优先读取 config.<env>.yaml，失败则回退到 config.yaml
		v.SetConfigName("config." + env)
		if err := v.ReadInConfig(); err != nil {
			v.SetConfigName("config")
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("read config failed: %w", err)
			}
		}
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

// MustLoadWithPath 指定路径读取配置（失败直接 panic）
func MustLoadWithPath(path string) *Config {
	cfg, err := LoadWithPath(path)
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
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.max_file_size_mb", 20)
	v.SetDefault("storage.max_total_size_mb", 5120)
	v.SetDefault("storage.local.base_dir", "./storage")
	v.SetDefault("storage.local.base_url", "http://localhost:8080/storage")
}
