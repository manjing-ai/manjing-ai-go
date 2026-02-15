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
	Email   EmailConfig   `mapstructure:"email"`
	Swagger SwaggerConfig `mapstructure:"swagger"`
	LLM     LLMConfig     `mapstructure:"llm"`
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

// LLMConfig LLM 大语言模型配置
type LLMConfig struct {
	Default LLMModelConfig `mapstructure:"default"`
}

// LLMModelConfig 单个模型配置
type LLMModelConfig struct {
	BaseURL     string  `mapstructure:"base_url"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float32 `mapstructure:"temperature"`
	Timeout     int     `mapstructure:"timeout"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Provider  string                      `mapstructure:"provider"` // tencent_smtp
	SMTP      EmailSMTPConfig             `mapstructure:"smtp"`
	FromName  string                      `mapstructure:"from_name"`
	FromAddr  string                      `mapstructure:"from_address"`
	Templates map[string]string           `mapstructure:"templates"`
	Subjects  map[string]string           `mapstructure:"subjects"`
	Scenes    map[string]EmailSceneConfig `mapstructure:"scenes"`
	Code      EmailCodeConfig             `mapstructure:"code"`
	RateLimit EmailRateLimitConfig        `mapstructure:"rate_limit"`
	Extra     map[string]interface{}      `mapstructure:"extra"`
}

// EmailSMTPConfig SMTP 配置
type EmailSMTPConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	UseSSL      bool   `mapstructure:"use_ssl"`
	UseStartTLS bool   `mapstructure:"use_starttls"`
}

// EmailSceneConfig 场景配置
type EmailSceneConfig struct {
	TemplateCode string `mapstructure:"template_code"`
	Subject      string `mapstructure:"subject"`
	TTLSeconds   int    `mapstructure:"ttl_seconds"`
}

// EmailCodeConfig 验证码配置
type EmailCodeConfig struct {
	TTLSeconds int `mapstructure:"ttl_seconds"`
	Length     int `mapstructure:"length"`
}

// EmailRateLimitConfig 频控配置
type EmailRateLimitConfig struct {
	IntervalSeconds int `mapstructure:"interval_seconds"`
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
	v.SetDefault("email.provider", "tencent_smtp")
	v.SetDefault("email.code.ttl_seconds", 300)
	v.SetDefault("email.code.length", 6)
	v.SetDefault("email.rate_limit.interval_seconds", 60)
	v.SetDefault("email.scenes.register.template_code", "EMAIL_REGISTER")
	v.SetDefault("email.scenes.register.subject", "验证码")
	v.SetDefault("email.scenes.register.ttl_seconds", 300)
	v.SetDefault("email.scenes.reset_password.template_code", "EMAIL_RESET")
	v.SetDefault("email.scenes.reset_password.subject", "验证码")
	v.SetDefault("email.scenes.reset_password.ttl_seconds", 300)
	v.SetDefault("email.scenes.login.template_code", "EMAIL_LOGIN")
	v.SetDefault("email.scenes.login.subject", "验证码")
	v.SetDefault("email.scenes.login.ttl_seconds", 300)
	v.SetDefault("llm.default.base_url", "https://api.deepseek.com/v1")
	v.SetDefault("llm.default.model", "deepseek-chat")
	v.SetDefault("llm.default.max_tokens", 4096)
	v.SetDefault("llm.default.temperature", 0.7)
	v.SetDefault("llm.default.timeout", 60)
}
