package rocketmq

// Config RocketMQ 配置
type Config struct {
	NameServer string `yaml:"name-server" json:"name-server"` // NameServer 地址，多个用逗号分隔
	AccessKey  string `yaml:"access-key" json:"access-key"`   // AccessKey
	SecretKey  string `yaml:"secret-key" json:"secret-key"`   // SecretKey
	Namespace  string `yaml:"namespace" json:"namespace"`      // 命名空间
	GroupName  string `yaml:"group-name" json:"group-name"`    // 生产者/消费者 组名
	RetryTimes int    `yaml:"retry-times" json:"retry-times"`  // 重试次数，默认 3
}