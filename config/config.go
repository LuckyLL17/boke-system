package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Log      LogConfig
	Cache    CacheConfig
	Worker   WorkerConfig
}

type ServerConfig struct {
	Port    int
	Mode    string
	BaseURL string
	WebDir  string
	LogDir  string
}

type DatabaseConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	SSLMode     string
	AutoMigrate bool
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	Expires     int
	ExpireHours int
}

type StorageConfig struct {
	Type        string
	Local       LocalStorageConfig
	MinIO       MinIOConfig
	LocalPath   string
	BaseURL     string
	MaxSize     int64
	ChunkSize   int64
	AllowedExts []string
}

type LocalStorageConfig struct {
	Path string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type LogConfig struct {
	Level    string
	Path     string
	Filename string
}

type CacheConfig struct {
	RSSTTL   int
	StatsTTL int
}

type WorkerConfig struct {
	RSSCacheInterval      int
	StatsAggregateInterval int
	CleanupInterval       int
}

var AppConfig *Config

func Load(configPath ...string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(configPath) > 0 && configPath[0] != "" {
		v.SetConfigFile(configPath[0])
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	AppConfig = &Config{}
	if err := v.Unmarshal(AppConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	AppConfig.applyDefaults()
	return AppConfig, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 { c.Server.Port = 8080 }
	if c.Server.Mode == "" { c.Server.Mode = "debug" }
	if c.Server.BaseURL == "" { c.Server.BaseURL = fmt.Sprintf("http://localhost:%d", c.Server.Port) }
	if c.Server.WebDir == "" { c.Server.WebDir = "./web" }
	if c.Server.LogDir == "" { c.Server.LogDir = "./logs" }
	if c.Database.Host == "" { c.Database.Host = "localhost" }
	if c.Database.Port == 0 { c.Database.Port = 5432 }
	if c.Database.SSLMode == "" { c.Database.SSLMode = "disable" }
	if !c.Database.AutoMigrate { c.Database.AutoMigrate = true }
	if c.JWT.Secret == "" { c.JWT.Secret = "change-me" }
	if c.JWT.ExpireHours == 0 { c.JWT.ExpireHours = c.JWT.Expires; if c.JWT.ExpireHours == 0 { c.JWT.ExpireHours = 24 } }
	if c.Storage.LocalPath == "" {
		if c.Storage.Local.Path != "" { c.Storage.LocalPath = c.Storage.Local.Path } else { c.Storage.LocalPath = "./storage" }
	}
	if c.Storage.BaseURL == "" { c.Storage.BaseURL = c.Server.BaseURL + "/storage" }
	if c.Storage.MaxSize == 0 { c.Storage.MaxSize = 524288000 }
	if c.Storage.ChunkSize == 0 { c.Storage.ChunkSize = 5242880 }
	if len(c.Storage.AllowedExts) == 0 { c.Storage.AllowedExts = []string{"mp3","m4a","wav","ogg","flac","aac"} }
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}
