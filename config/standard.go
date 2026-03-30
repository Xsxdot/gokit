package config

// StandardConfig 框架基础设施标准配置聚合根
// 业务项目可通过结构体嵌套与此配置打平合并：
//
//	type AppConfig struct {
//	    config.StandardConfig `yaml:",inline" json:",inline"`
//	    // 业务专属配置...
//	    JWTSecret string `yaml:"jwt_secret"`
//	}
type StandardConfig struct {
	BaseInfo     BaseInfo           `yaml:"app" json:"app"`
	Database     Database           `yaml:"database" json:"database"`
	Redis        RedisConfig        `yaml:"redis" json:"redis"`
	Proxy        ProxyConfig        `yaml:"proxy" json:"proxy"`
	GRPC         GRPCConfig         `yaml:"grpc" json:"grpc"`
	OSS          OssConfig          `yaml:"oss" json:"oss"`
	Log          LogConfig          `yaml:"log" json:"log"`
	JWT          JwtConfig          `yaml:"jwt" json:"jwt"`
	ConfigCenter ConfigCenterConfig `yaml:"config_center" json:"config_center"`
}

type BaseInfo struct {
	AppName  string `yaml:"app-name"`
	Env      string `yaml:"env"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Domain   string `yaml:"domain"`
	LogLevel string `yaml:"log-level"`
}
