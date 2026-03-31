package config

// Database 数据库配置结构体（纯净定义，无依赖）
type Database struct {
	Host     string `yaml:"host" json:"host,omitempty"`
	Port     int64  `yaml:"port" json:"port,omitempty"`
	User     string `yaml:"user" json:"user,omitempty"`
	Password string `yaml:"password" json:"password,omitempty"`
	DbName   string `yaml:"db-name" json:"db-name,omitempty"`
}
