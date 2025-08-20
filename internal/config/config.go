// Package config internal/config/config.go
package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Storage  StorageConfig
	WebRTC   WebRTCConfig
	Security SecurityConfig
	Limits   LimitsConfig
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Environment     string        `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	SSLMode     string `mapstructure:"ssl_mode"`
	MaxConns    int32  `mapstructure:"max_conns"`
	MinConns    int32  `mapstructure:"min_conns"`
	MaxConnAge  string `mapstructure:"max_conn_age"`
	MaxIdleTime string `mapstructure:"max_idle_time"`
}

type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	OTPExpiry       time.Duration `mapstructure:"otp_expiry"`
	Secret          string        `mapstructure:"secret"`
	Issuer          string        `mapstructure:"issuer"`
}

type StorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

type WebRTCConfig struct {
	STUNServers []string `mapstructure:"stun_servers"`
	TURNServers []struct {
		URLs       []string `mapstructure:"urls"`
		Username   string   `mapstructure:"username"`
		Credential string   `mapstructure:"credential"`
	} `mapstructure:"turn_servers"`
	SFUEnabled bool `mapstructure:"sfu_enabled"`
}

type SecurityConfig struct {
	BCryptCost      int      `mapstructure:"bcrypt_cost"`
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
	TrustedProxies  []string `mapstructure:"trusted_proxies"`
	RateLimitWindow string   `mapstructure:"rate_limit_window"`
}

type LimitsConfig struct {
	MaxGroupMembers     int   `mapstructure:"max_group_members"`
	MaxMessageSize      int64 `mapstructure:"max_message_size"`
	MaxAttachmentSize   int64 `mapstructure:"max_attachment_size"`
	MessageRatePerMin   int   `mapstructure:"message_rate_per_min"`
	CallMaxParticipants int   `mapstructure:"call_max_participants"`
}

func Load() (*Config, error) {
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.read_timeout", 15*time.Second)
	viper.SetDefault("server.write_timeout", 15*time.Second)
	viper.SetDefault("auth.access_token_ttl", 15*time.Minute)
	viper.SetDefault("auth.refresh_token_ttl", 7*24*time.Hour)
	viper.SetDefault("limits.max_group_members", 512)
	viper.SetDefault("security.bcrypt_cost", 12)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
