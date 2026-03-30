package utils

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

const (
	// 起始时间戳 (2024-01-01 00:00:00 UTC)，可用约 69 年
	twepoch = int64(1704067200000)

	// 位数分配
	workerIDBits     = uint(5) // 机器 ID 位数
	datacenterIDBits = uint(5) // 数据中心 ID 位数
	sequenceBits     = uint(12) // 序列号位数

	// 最大值计算
	maxWorkerID     = int64(-1 ^ (-1 << workerIDBits))     // 31
	maxDatacenterID = int64(-1 ^ (-1 << datacenterIDBits)) // 31
	maxSequence     = int64(-1 ^ (-1 << sequenceBits))     // 4095

	// 位移
	workerIDShift      = sequenceBits                       // 12
	datacenterIDShift  = sequenceBits + workerIDBits        // 17
	timestampLeftShift = sequenceBits + workerIDBits + datacenterIDBits // 22
)

var (
	// ErrInvalidWorkerID 机器 ID 无效
	ErrInvalidWorkerID = errors.New("worker ID must be between 0 and 31")
	// ErrInvalidDatacenterID 数据中心 ID 无效
	ErrInvalidDatacenterID = errors.New("datacenter ID must be between 0 and 31")
	// ErrSystemClock 时钟回拨错误
	ErrSystemClock = errors.New("clock moved backwards")
)

// Snowflake 雪花算法 ID 生成器
type Snowflake struct {
	mu            sync.Mutex
	datacenterID  int64
	workerID      int64
	sequence      int64
	lastTimestamp int64
}

// 默认实例
var defaultSnowflake *Snowflake
var once sync.Once

// InitSnowflake 初始化默认雪花算法实例
// datacenterID: 数据中心 ID (0-31)
// workerID: 机器 ID (0-31)
func InitSnowflake(datacenterID, workerID int64) error {
	sf, err := NewSnowflake(datacenterID, workerID)
	if err != nil {
		return err
	}
	defaultSnowflake = sf
	return nil
}

// NewSnowflake 创建雪花算法实例
func NewSnowflake(datacenterID, workerID int64) (*Snowflake, error) {
	if datacenterID < 0 || datacenterID > maxDatacenterID {
		return nil, ErrInvalidDatacenterID
	}
	if workerID < 0 || workerID > maxWorkerID {
		return nil, ErrInvalidWorkerID
	}
	return &Snowflake{
		datacenterID:  datacenterID,
		workerID:      workerID,
		sequence:      0,
		lastTimestamp: -1,
	}, nil
}

// NextID 生成下一个 ID
func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// 时钟回拨检测
	if now < s.lastTimestamp {
		return 0, ErrSystemClock
	}

	// 同一毫秒内
	if now == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 序列号溢出，等待下一毫秒
		if s.sequence == 0 {
			for now <= s.lastTimestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 不同毫秒，序列号重置
		s.sequence = 0
	}

	s.lastTimestamp = now

	// 组装 ID
	id := ((now - twepoch) << timestampLeftShift) |
		(s.datacenterID << datacenterIDShift) |
		(s.workerID << workerIDShift) |
		s.sequence

	return id, nil
}

// NextIDDefault 使用默认实例生成 ID（需要先调用 InitSnowflake 初始化）
func NextIDDefault() (int64, error) {
	if defaultSnowflake == nil {
		// 自动初始化，使用默认值 0,0
		once.Do(func() {
			defaultSnowflake, _ = NewSnowflake(0, 0)
		})
	}
	return defaultSnowflake.NextID()
}

// NextIDString 生成字符串形式的 ID
func NextIDString() (string, error) {
	id, err := NextIDDefault()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// ParseID 解析 ID 获取各组成部分
func ParseID(id int64) (timestamp int64, datacenterID int64, workerID int64, sequence int64) {
	sequence = id & maxSequence
	workerID = (id >> workerIDShift) & maxWorkerID
	datacenterID = (id >> datacenterIDShift) & maxDatacenterID
	timestamp = (id >> timestampLeftShift) + twepoch
	return
}