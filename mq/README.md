# 消息队列组件 (mq)

该组件提供了一套统一的消息队列抽象接口，支持多种后端驱动（目前支持 Redis Stream 和 RocketMQ）。通过拆分到独立子包，实现按需引入依赖。

## 包结构

| 包路径 | 说明 | 引入的依赖 |
|--------|------|-----------|
| `github.com/xsxdot/gokit/mq` | 公共接口定义 | 无额外依赖 |
| `github.com/xsxdot/gokit/mq/redis` | Redis Stream 实现 | `go-redis/v9` |
| `github.com/xsxdot/gokit/mq/rocketmq` | RocketMQ 实现 | `rocketmq-client-go/v2` |

## 特性

- **统一接口**: 屏蔽底层实现差异，提供统一的 `Producer` 和 `Consumer` 接口
- **依赖隔离**: 只导入需要的子包，避免依赖爆炸
- **多驱动支持**:
    - **Redis Stream**: 适合轻量级、低延迟场景，复用外部 Redis 客户端
    - **RocketMQ**: 适合高吞吐、分布式事务和严格的消息可靠性场景
- **消息模式**:
    - **普通消息**: 即时发送与消费
    - **延时消息**: 
        - RocketMQ: 使用原生延时级别（18个预设值）
        - Redis: 通过 Sorted Set + 定时轮询实现任意时长延时
    - **顺序消息**:
        - RocketMQ: 通过 ShardingKey 路由，完整支持
        - Redis: 通过扩展 Topic 名称实现分区保序（有功能限制）

## 快速开始

### Redis Stream

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/xsxdot/gokit/mq"
    "github.com/xsxdot/gokit/mq/redis"
)

// 初始化 Redis 客户端（或通过 RDBComponent）
rdb := redis.NewClient(&redis.Options{
    Addr: "127.0.0.1:6379",
})

cfg := redis.Config{
    Client: rdb,  // 必填：Redis 客户端
}

// 创建生产者
producer, err := redis.NewProducer(&cfg)

// 创建消费者
consumer, err := redis.NewConsumer(&cfg)
```

### RocketMQ

```go
import (
    "github.com/xsxdot/gokit/mq"
    "github.com/xsxdot/gokit/mq/rocketmq"
)

cfg := rocketmq.Config{
    NameServer: "127.0.0.1:9876",
    GroupName:  "my-group",
    RetryTimes: 3,
}

// 创建生产者
producer, err := rocketmq.NewProducer(&cfg)

// 创建消费者
consumer, err := rocketmq.NewConsumer(&cfg)
```

## 使用示例

### 发送消息

```go
// 普通消息
result, err := producer.SendMessage(ctx, &mq.Message{
    Topic: "order-created",
    Body:  []byte(`{"order_id": "12345"}`),
    Key:   "order-12345",
})

// 延时消息（30分钟后执行）
result, err := producer.SendDelayMessage(ctx, &mq.Message{
    Topic: "order-cancel-check",
    Body:  []byte(`{"order_id": "12345"}`),
}, 30*time.Minute)

// 顺序消息（相同 shardingKey 的消息按顺序消费）
result, err := producer.SendOrderMessage(ctx, &mq.Message{
    Topic: "order-status-update",
    Body:  []byte(`{"order_id": "12345"}`),
}, "order-12345")
```

### 消费消息

```go
// 订阅主题（集群模式）
mq.SubscribeCluster(consumer, "order-created", "order-group",
    func(ctx context.Context, msg *mq.Message) mq.ConsumeResult {
        fmt.Printf("收到消息: %s\n", string(msg.Body))
        return mq.ConsumeSuccess
    })

// 启动消费者
consumer.Start()
defer consumer.Close()
```

## Component 方式

### Redis MQ Component

```go
var producer mq.Producer
var consumer mq.Consumer

cfg := redis.Config{
    Client: rdb,
}

comp := redis.NewComponent("mq.redis").
    WithProducer(&producer).
    WithConsumer(&consumer)

comp.Start(ctx, &cfg)
defer comp.Stop()
```

### RocketMQ Component

```go
var producer mq.Producer
var consumer mq.Consumer

comp := rocketmq.NewComponent("mq.rocketmq").
    WithProducer(&producer).
    WithConsumer(&consumer)

comp.Start(ctx, &cfg)
defer comp.Stop()
```

## 配置项详解

### Redis 配置 (`redis.Config`)

| 字段 | 类型 | 说明 |
|------|------|------|
| `Client` | `*redis.Client` | Redis 客户端（必填） |
| `DelayPollIntervalMs` | `int` | 延时消息检查间隔（毫秒），默认 500 |
| `BlockTimeoutMs` | `int` | 消费者拉取阻塞时长（毫秒），默认 5000 |
| `PendingRetryIntervalSec` | `int` | Pending 消息重试间隔（秒），默认 30 |

### RocketMQ 配置 (`rocketmq.Config`)

| 字段 | 类型 | 说明 |
|------|------|------|
| `NameServer` | `string` | NameServer 地址，多个用逗号分隔 |
| `AccessKey` | `string` | AccessKey（可选） |
| `SecretKey` | `string` | SecretKey（可选） |
| `Namespace` | `string` | 命名空间（可选） |
| `GroupName` | `string` | 生产者/消费者组名 |
| `RetryTimes` | `int` | 重试次数，默认 3 |

## 驱动特性对比

| 特性 | Redis | RocketMQ |
|------|-------|----------|
| 普通消息 | ✅ | ✅ |
| 延时消息 | ✅ 任意时长 | ✅ 18个预设级别 |
| 顺序消息 | ⚠️ 有严重限制 | ✅ 完整支持 |
| 广播消费 | ⚠️ 有功能限制 | ✅ 完整支持 |
| 连接管理 | 复用外部客户端 | 内部管理 |
| 生产环境 | 适合中小规模 | 适合大规模 |

## 注意事项

1. **Redis 客户端管理**: Redis MQ 不创建独立连接，需用户提供客户端（通过 RDBComponent）
2. **Redis 顺序消息限制**: 存在设计缺陷，消息发送到 `{topic}:order:{shardingKey}`，但消费者只订阅 `{topic}`，导致无法消费。如需顺序消息，请使用 RocketMQ
3. **Redis 广播消费限制**: 无重试机制、新消费者丢失历史消息、无 ACK 机制
4. **RocketMQ 延时级别**: 仅支持固定级别（1s, 5s, 10s, 30s, 1m, 2m, 3m, 4m, 5m, 6m, 7m, 8m, 9m, 10m, 20m, 30m, 1h, 2h）
5. **依赖隔离**: 只导入需要的子包，避免引入不必要的依赖