package config

// ProxyConfig SOCKS代理配置结构体（纯净定义，无依赖）
type ProxyConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`   // 是否启用代理
	Host     string `yaml:"host" json:"host"`         // 代理服务器地址
	Port     int    `yaml:"port" json:"port"`         // 代理服务器端口
	Username string `yaml:"username" json:"username"` // 代理认证用户名（可选）
	Password string `yaml:"password" json:"password"` // 代理认证密码（可选）
}