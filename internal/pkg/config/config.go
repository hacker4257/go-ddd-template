package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	App  AppConfig  `koanf:"app"`
	HTTP HTTPConfig `koanf:"http"`
	Log  LogConfig  `koanf:"log"`
	DB   DBConfig   `koanf:"db"`
	Redis RedisConfig `koanf:"redis"`
	Kafka KafkaConfig `koanf:"kafka"`

}

type KafkaConfig struct {
	Brokers   []string `koanf:"brokers"`
	UserTopic string   `koanf:"user_topic"`
}


type RedisConfig struct {
	Addr     string        `koanf:"addr"`
	Password string        `koanf:"password"`
	DB       int           `koanf:"db"`
	UserTTL  time.Duration `koanf:"user_ttl"`
}


type AppConfig struct {
	Name string `koanf:"name"`
	Env  string `koanf:"env"`
}

type DBConfig struct {
	MySQL MySQLConfig `koanf:"mysql"`
}

type MySQLConfig struct {
	DSN             string        `koanf:"dsn"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
}


type HTTPConfig struct {
	Addr         string        `koanf:"addr"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	IdleTimeout  time.Duration `koanf:"idle_timeout"`
}

type LogConfig struct {
	Level string `koanf:"level"` // debug/info/warn/error
}

func Load(path string) (Config, error) {
	k := koanf.New(".")

	// 1) file config
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("load config file: %w", err)
	}

	// 2) env override: APP_NAME, HTTP_ADDR, LOG_LEVEL, etc.
	// example: HTTP_ADDR=":8081" will override http.addr
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil); err != nil {
		return Config{}, fmt.Errorf("load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	// defaults
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.HTTP.ReadTimeout == 0 {
		cfg.HTTP.ReadTimeout = 5 * time.Second
	}
	if cfg.HTTP.WriteTimeout == 0 {
		cfg.HTTP.WriteTimeout = 10 * time.Second
	}
	if cfg.HTTP.IdleTimeout == 0 {
		cfg.HTTP.IdleTimeout = 60 * time.Second
	}
	if cfg.App.Name == "" {
		cfg.App.Name = "go-ddd-template"
	}
	if cfg.App.Env == "" {
		cfg.App.Env = "dev"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	//mysql
	if cfg.DB.MySQL.MaxOpenConns == 0 { 
		cfg.DB.MySQL.MaxOpenConns = 10 
	}

	if cfg.DB.MySQL.MaxIdleConns == 0 { 
		cfg.DB.MySQL.MaxIdleConns = 5 
	}

	if cfg.DB.MySQL.ConnMaxLifetime == 0 { 
		cfg.DB.MySQL.ConnMaxLifetime = 30 * time.Minute 
	}

	//redis
	if cfg.Redis.Addr == "" { 
		cfg.Redis.Addr = "127.0.0.1:6379" 
	}

	if cfg.Redis.UserTTL == 0 { 
		cfg.Redis.UserTTL = 10 * time.Minute 
	}

	//kafka
	if len(cfg.Kafka.Brokers) == 0 { 
		cfg.Kafka.Brokers = []string{"127.0.0.1:9092"} 
	}
	if cfg.Kafka.UserTopic == "" { 
		cfg.Kafka.UserTopic = "user.events" 
	}


	return cfg, nil
}
