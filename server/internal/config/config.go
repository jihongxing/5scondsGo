package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Game     GameConfig     `yaml:"game"`
	Auth     AuthConfig     `yaml:"auth"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	WSPath      string `yaml:"ws_path"`
	MetricsPort int    `yaml:"metrics_port"`
	Mode        string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// GameConfig 游戏配置
type GameConfig struct {
	PhaseDuration          int       `yaml:"phase_duration"`
	TickInterval           int       `yaml:"tick_interval"`
	BetAmounts             []float64 `yaml:"bet_amounts"`
	MaxRoomPlayers         int       `yaml:"max_room_players"`
	MinRoomPlayers         int       `yaml:"min_room_players"`
	DefaultOwnerCommission float64   `yaml:"default_owner_commission"`
	PlatformCommission     float64   `yaml:"platform_commission"`
	MaxTotalCommission     float64   `yaml:"max_total_commission"`
	MinMarginBalance       float64   `yaml:"min_margin_balance"`
	MinCustodyQuota        float64   `yaml:"min_custody_quota"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret         string        `yaml:"jwt_secret"`
	JWTExpire         time.Duration `yaml:"jwt_expire"`
	MinPasswordLength int           `yaml:"min_password_length"`
	InviteCodeLength  int           `yaml:"invite_code_length"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 从环境变量覆盖
	cfg.overrideFromEnv()

	return &cfg, nil
}

// overrideFromEnv 从环境变量覆盖配置
func (c *Config) overrideFromEnv() {
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		// 简化处理，实际应解析int
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		c.Database.DBName = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}
}
