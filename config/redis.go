package config

// RedisConfig Redis 配置结构体（纯净定义，无依赖）
type RedisConfig struct {
	Mode     string `yaml:"mode"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}